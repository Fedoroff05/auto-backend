package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fedoroff05/auto-backend/config"
	deliveryHttp "github.com/Fedoroff05/auto-backend/internal/handler/http"
	v1 "github.com/Fedoroff05/auto-backend/internal/handler/http/v1"
	"github.com/Fedoroff05/auto-backend/internal/repository/postgres"
	"github.com/Fedoroff05/auto-backend/internal/usecase"
	"github.com/Fedoroff05/auto-backend/pkg/hasher"
	"github.com/Fedoroff05/auto-backend/pkg/jwt"
	pkgPostgres "github.com/Fedoroff05/auto-backend/pkg/postgres"
)

// @title           Auto Service Backend API
// @version         1.0
// @description     REST API для сервиса объявлений по продаже автомобилей.
// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Токен доступа в формате: Bearer <token>
func main() {
	//инициализация конфигурации
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Println("configuration successfully loaded")

	//инициализация пула соединений бд
	ctx := context.Background()
	pgPool, err := pkgPostgres.New(ctx, pkgPostgres.Config{
		DSN:             cfg.Postgres.DSN(),
		MaxConns:        cfg.Postgres.MaxConns,
		MinConns:        cfg.Postgres.MinConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
		MaxConnIdleTime: cfg.Postgres.MaxConnIdleTime,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pgPool.Close()
	log.Println("database connection pool initialized")

	//инициализация пакетов вспомогательных
	passwordHasher := hasher.NewBcryptHasher(10)

	tokenManager, err := jwt.NewTokenManager(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)
	if err != nil {
		log.Fatalf("Failed to init token manager: %v", err)
	}

	userRepo := postgres.NewUserRepository(pgPool)

	authUsecase := usecase.NewAuthUsecase(userRepo, passwordHasher, tokenManager)

	authHandler := v1.NewAuthHandler(authUsecase)
	router := deliveryHttp.NewRouter(authHandler, tokenManager)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		log.Printf("HTTP server is listening on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to listen and serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	log.Printf("received signal: %v. initiating graceful shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown with error: %v", err)
	}

	log.Println("server exited gracefully")
}
