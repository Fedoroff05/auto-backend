package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Fedoroff05/auto-backend/internal/domain"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/response"
)

type AuthHandler struct {
	authUsecase domain.UserUsecase
}

func NewAuthHandler(authUsecase domain.UserUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// DTO
type registerRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Name     string  `json:"name"`
	Phone    *string `json:"phone,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создает учетную запись пользователя с ролью user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body registerRequest true "Параметры регистрации"
// @Success      201 {object} response.Response{data=domain.User}
// @Failure      400 {object} response.ErrorResponse
// @Failure      409 {object} response.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.Error(w, http.StatusBadRequest, "email, password and name are required")
		return
	}
	user, err := h.authUsecase.Register(r.Context(), req.Email, req.Password, req.Name, req.Phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация по email и паролю с получением пары JWT токенов
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body loginRequest true "Учетные данные"
// @Success      200 {object} response.Response{data=authResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	accessToken, refreshToken, err := h.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to authenticate")
		return
	}
	response.JSON(w, http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshToken godoc
// @Summary      Обновление JWT токенов
// @Description  Выпуск новой пары access/refresh токенов по действующему refresh токену
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body refreshRequest true "Refresh токен"
// @Success      200 {object} response.Response{data=authResponse}
// @Failure      400 {object} response.ErrorResponse
// @Failure      401 {object} response.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}
	newAccessToken, newRefreshToken, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenExpired) || errors.Is(err, domain.ErrInvalidRefreshToken) {
			response.Error(w, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to refresh token")
		return
	}
	response.JSON(w, http.StatusOK, authResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

// GetProfile godoc
// @Summary      Получение профиля текущего пользователя
// @Description  Возвращает профиль аутентифицированного пользователя
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=domain.User}
// @Failure      401 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Router       /auth/profile [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.authUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get profile")
		return
	}
	response.JSON(w, http.StatusOK, user)
}
