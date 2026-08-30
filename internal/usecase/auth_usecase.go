package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fedoroff05/auto-backend/internal/domain"
	"github.com/Fedoroff05/auto-backend/pkg/hasher"
	"github.com/Fedoroff05/auto-backend/pkg/jwt"
	"github.com/google/uuid"
	"strings"
	"time"
)

// реализует бизнес-сценарии аутентификакции и работы с профилем
type AuthUsecase struct {
	userRepo     domain.UserRepository
	hasher       hasher.PasswordHasher
	tokenManager *jwt.TokenManager
}

// конструктор юзкейса с внедрением зависимостей
func NewAuthUsecase(
	userRepo domain.UserRepository,
	hasher hasher.PasswordHasher,
	tokenManager *jwt.TokenManager,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
	}
}

// регистрация нового пользователя
func (u *AuthUsecase) Register(ctx context.Context, email, password, name string, phone *string) (*domain.User, error) {
	//нормализация входных данных
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanName := strings.TrimSpace(name)

	//хэширование пароля через bcrypt
	hashedPassword, err := u.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("auth_usecase: failed to hash password: %w", err)
	}

	//формирование доменной модели
	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        cleanEmail,
		PasswordHash: hashedPassword,
		Name:         cleanName,
		Phone:        phone,
		Role:         domain.RoleUser, //новые пользователи всегда получают базовую роль
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	//сохранение через интерфейс репозитория
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("auth_usecase: failed to create user: %w", err)
	}
	return user, nil
}

// проверяет уч.данные и генерирует пару access и refresh токенов
func (u *AuthUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	//поиск пользователя по почте
	user, err := u.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", "", domain.ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("auth_usecase: failed to find user: %w", err)
	}

	//сверка хэша пароля
	if !u.hasher.Compare(password, user.PasswordHash) {
		return "", "", domain.ErrInvalidCredentials
	}

	//генерация access токена
	accessToken, err := u.tokenManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("auth_usecase: failed to generate access token: %w", err)
	}

	//генерация refresh токена
	refreshToken, err := u.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("auth_usecase: failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// возвращает данные юзера по его id
func (u *AuthUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth_usecase: failed to get user profile: %w", err)
	}
	return user, nil
}

// проверяет старый refresh и выпускает новую пару токенов
func (u *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	//валидация подписи и срока действия refresh токена
	userID, err := u.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return "", "", domain.ErrRefreshTokenExpired
		}
		return "", "", domain.ErrInvalidRefreshToken
	}

	//проверка, существует ли еще юзер в БД и получаем актуальную роль
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", "", domain.ErrInvalidRefreshToken
		}
		return "", "", fmt.Errorf("auth_usecase: failed to check user on refresh: %w", err)
	}

	//заливается новая пара токенов
	newAccessToken, err := u.tokenManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("auth_usecase: failed to generate access token: %w", err)
	}

	newRefreshToken, err := u.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("auth_usecase: failed to generate refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}
