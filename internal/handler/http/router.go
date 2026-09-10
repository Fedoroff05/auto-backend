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

func NewRouter(
	authHandler *v1.AuthHandler,
	listingHandler *v1.ListingHandler,
	valuationHandler *v1.ValuationHandler,
	favoriteHandler *v1.FavoriteHandler,
	chatHandler *v1.ChatHandler,
	abHandler *v1.ABHandler,
	tokenManager *jwt.TokenManager,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	//WebSocket роут (эндпоинт /ws?token=...)
	r.Get("/ws", chatHandler.ConnectWS)

	r.Route("/api/v1", func(r chi.Router) {
		//auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(tokenManager))
				r.Get("/profile", authHandler.GetProfile)
			})
		})

		//catalog
		r.Get("/brands", listingHandler.GetBrands)
		r.Get("/brands/{brand_id}/models", listingHandler.GetModels)

		//listings
		r.Route("/listings", func(r chi.Router) {
			r.Get("/", listingHandler.List)
			r.Get("/{id}", listingHandler.GetByID)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(tokenManager))
				r.Post("/", listingHandler.Create)
				r.Put("/{id}", listingHandler.Update)
				r.Delete("/{id}", listingHandler.Delete)
				r.Post("/{id}/images", listingHandler.UploadImage)
			})
		})

		//valuation
		r.Get("/valuation/estimate", valuationHandler.EstimatePrice)

		//favorites
		r.Route("/favorites", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tokenManager))
			r.Get("/", favoriteHandler.GetUserFavorites)
			r.Post("/{listing_id}", favoriteHandler.Add)
			r.Delete("/{listing_id}", favoriteHandler.Remove)
		})

		//chats
		r.Route("/chats", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tokenManager))
			r.Get("/", chatHandler.GetUserChats)
			r.Post("/start", chatHandler.StartChat)
			r.Get("/{chat_id}/messages", chatHandler.GetMessages)
		})

		//AB testing
		r.Route("/ab", func(r chi.Router) {
			r.Post("/events", abHandler.TrackEvent)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(tokenManager))
				r.Get("/variant", abHandler.GetVariant)
			})
		})
	})

	return r
}
