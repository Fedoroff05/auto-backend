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
	pkgS3 "github.com/Fedoroff05/auto-backend/pkg/s3"
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
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Configuration loaded")

	ctx := context.Background()

	pgPool, err := pkgPostgres.New(ctx, pkgPostgres.Config{
		DSN:             cfg.Postgres.DSN(),
		MaxConns:        cfg.Postgres.MaxConns,
		MinConns:        cfg.Postgres.MinConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
		MaxConnIdleTime: cfg.Postgres.MaxConnIdleTime,
	})
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()
	log.Println("PostgreSQL connection pool initialized")

	s3Client, err := pkgS3.NewClient(ctx, pkgS3.Config{
		Endpoint:        cfg.MinIO.Endpoint,
		AccessKeyID:     cfg.MinIO.RootUser,
		SecretAccessKey: cfg.MinIO.RootPassword,
		BucketName:      cfg.MinIO.BucketName,
		UseSSL:          cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 MinIO client: %v", err)
	}
	log.Println("MinIO S3 client initialized and bucket verified")

	passwordHasher := hasher.NewBcryptHasher(10)
	tokenManager, err := jwt.NewTokenManager(cfg.JWT.SecretKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	if err != nil {
		log.Fatalf("Failed to init token manager: %v", err)
	}
	userRepo := postgres.NewUserRepository(pgPool)
	listingRepo := postgres.NewListingRepository(pgPool)
	authUsecase := usecase.NewAuthUsecase(userRepo, passwordHasher, tokenManager)
	listingUsecase := usecase.NewListingUsecase(listingRepo, s3Client)
	authHandler := v1.NewAuthHandler(authUsecase)
	listingHandler := v1.NewListingHandler(listingUsecase)
	router := deliveryHttp.NewRouter(authHandler, listingHandler, tokenManager)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		log.Printf("HTTP Server is listening on port %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server stopped: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
