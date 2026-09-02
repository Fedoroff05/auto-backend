package http

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	v1 "github.com/Fedoroff05/auto-backend/internal/handler/http/v1"
	"github.com/Fedoroff05/auto-backend/pkg/jwt"
)

func NewRouter(authHandler *v1.AuthHandler, tokenManager *jwt.TokenManager) *chi.Mux {
	r := chi.NewRouter()

	//база
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	//API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(tokenManager))
				r.Get("/profile", authHandler.GetProfile)
			})
		})
	})
	return r
}
