package server

import (
	"database/sql"
	"web_request_methods/handlers"
	"web_request_methods/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

type Server struct {
	db     *sql.DB
	router *chi.Mux
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		db:     db,
		router: chi.NewRouter(),
	}
}

func (s *Server) InitRouter() *chi.Mux {
	// CORS handler with support for multiple origins
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://127.0.0.1:5500",
			"http://localhost:5500",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept", "Authorization", "Content-Type", "X-CSRF-Token",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Apply middleware
	s.router.Use(corsHandler.Handler)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.AllowContentType("application/json"))

	// Initialize routes
	userHandler := handlers.NewUserHandler(s.db)
	
	s.router.Route("/api/users", func(r chi.Router) {
		// Public routes
		r.Post("/", userHandler.CreateUser)
		r.Get("/", userHandler.GetUsers)
		r.Get("/{id}", userHandler.GetUser)
		r.Put("/{id}", userHandler.UpdateUser)
		r.Delete("/{id}", userHandler.DeleteUser)
		r.Post("/login", userHandler.Login)

		// Protected routes: Require JWT Authorization
		r.With(middlewares.JWTAuth).Get("/profile", userHandler.GetUserProfile)
	})

	productHandler := handlers.NewProductHandler(s.db)

	s.router.Route("/api/products", func(r chi.Router) {
		r.Post("/", productHandler.CreateProduct)
		r.Get("/", productHandler.GetProducts)
		r.Get("/{id}", productHandler.GetProduct)
		r.Put("/{id}", productHandler.UpdateProduct)
		r.Delete("/{id}", productHandler.DeleteProduct)
	})

	return s.router
}
