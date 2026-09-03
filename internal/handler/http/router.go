package http

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/Fedoroff05/auto-backend/docs"
	"github.com/Fedoroff05/auto-backend/internal/handler/http/middleware"
	v1 "github.com/Fedoroff05/auto-backend/internal/handler/http/v1"
	"github.com/Fedoroff05/auto-backend/pkg/jwt"
)

func NewRouter(authHandler *v1.AuthHandler, tokenManager *jwt.TokenManager) *chi.Mux {
	r := chi.NewRouter()

	//база
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	//настройка CORS для взаимодействия с фронтендом
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, //кэширование preflight-запросов на 5 минут
	}))

	//эндпоинт для Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	//маршруты API
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
