package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yusuf4ktas/backend-project/internal/config"
	"github.com/yusuf4ktas/backend-project/internal/service"
)

type Server struct {
	config             *config.Config
	logger             *slog.Logger
	jwtSecret          []byte
	router             http.Handler
	userService        service.UserService
	userHandler        *UserHandler
	transactionHandler *TransactionHandler
	authHandler        *AuthHandler
	balanceHandler     *BalanceHandler
}

func NewServer(config *config.Config, logger *slog.Logger, userService service.UserService, userHandler *UserHandler, txHandler *TransactionHandler, authHandler *AuthHandler, balanceHandler *BalanceHandler) *Server { // MODIFIED
	s := &Server{
		config:             config,
		logger:             logger,
		userService:        userService,
		userHandler:        userHandler,
		transactionHandler: txHandler,
		authHandler:        authHandler,
		balanceHandler:     balanceHandler,
		jwtSecret:          []byte(config.JWTSecret),
	}
	s.router = s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() http.Handler {
	router := chi.NewRouter()

	//CORS header setup
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Any path like frontend etc. can be added to AllowedOrigins.
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler)

	router.Use(s.RequestLogger)
	router.Use(s.PrometheusMiddleware)
	router.Use(RateLimiter)
	router.Use(middleware.Recoverer)

	// --- Public Routes ---
	router.Get("/metrics", promhttp.Handler().ServeHTTP)
	router.Post("/api/v1/auth/register", appHandler(s.userHandler.Register).ServeHTTP)
	router.Post("/api/v1/auth/login", appHandler(s.authHandler.Login).ServeHTTP)

	// --- Protected Routes ---
	// All routes in this group require a valid token (AuthMiddleware).
	router.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)

		// Routes for any authenticated user
		r.Get("/api/v1/users/{id}", appHandler(s.userHandler.GetUserByID).ServeHTTP)
		r.Delete("/api/v1/users/{id}", appHandler(s.userHandler.DeleteUser).ServeHTTP)
		r.Post("/api/v1/transactions/transfer", appHandler(s.transactionHandler.Transfer).ServeHTTP)
		r.Get("/api/v1/transactions/history", appHandler(s.transactionHandler.GetTransactionHistory).ServeHTTP)
		r.Get("/api/v1/transactions/{id}", appHandler(s.transactionHandler.GetByTransactionID).ServeHTTP)
		r.Get("/api/v1/balances/current", appHandler(s.balanceHandler.GetCurrentBalance).ServeHTTP)

		// --- Admin-Only Routes ---
		// Require both a valid token and admin privileges.
		r.Group(func(r chi.Router) {
			r.Use(s.AdminOnlyMiddleware)

			r.Get("/api/v1/users", appHandler(s.userHandler.GetAllUsers).ServeHTTP)
			r.Post("/api/v1/transactions/credit", appHandler(s.transactionHandler.Credit).ServeHTTP)
			r.Post("/api/v1/transactions/debit", appHandler(s.transactionHandler.Debit).ServeHTTP)
		})
	})

	return router
}

// Getter method for the Router
func (s *Server) Router() http.Handler {
	return s.router
}
