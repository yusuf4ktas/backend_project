package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql" // The MySQL driver
	"github.com/redis/go-redis/v9"
	"github.com/yusuf4ktas/backend-project/internal/config"
	"github.com/yusuf4ktas/backend-project/internal/logger"
	"github.com/yusuf4ktas/backend-project/internal/repository"
	"github.com/yusuf4ktas/backend-project/internal/server"
	"github.com/yusuf4ktas/backend-project/internal/service"
	"github.com/yusuf4ktas/backend-project/internal/worker"
)

func main() {
	cfg, error := config.LoadConfig()
	if error != nil {
		fmt.Fprintf(os.Stderr, "FATAL: failed to load configuration: %v\n", error)
		os.Exit(1)
	}

	// --- Logger Setup ---
	log := logger.CreateLogger(cfg.Env)
	log.Info("logger initialized successfully")

	// --- Database Connection ---
	db, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		log.Error("could not connect to database", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		log.Error("could not ping database", "error", err)
		os.Exit(1)
	}
	log.Info("Database connection established successfully.")

	// --- Redis Connection --- //
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Error("could not connect to redis", "error", err)
		os.Exit(1)
	}
	log.Info("Redis connection established successfully.")

	// --- Dependency Injection ---
	userRepo := repository.NewUserRepository(db, rdb)
	balanceRepo := repository.NewBalanceRepository(db, rdb)
	transactionRepo := repository.NewTransactionRepository(db, rdb)
	auditRepo := repository.NewAuditLogRepository(db)

	auditService := service.NewAuditLogService(auditRepo)
	userService := service.NewUserService(userRepo, auditService, balanceRepo)
	transactionService := service.NewTransactionService(db, rdb, transactionRepo, balanceRepo, auditService)
	balanceService := service.NewBalanceService(balanceRepo)

	// ---  Worker Pool Setup ---
	dispatcher := worker.NewDispatcher(5, transactionService)
	dispatcher.Run(context.Background())
	log.Info("Worker pool started.")

	// --- Handlers and Server Setup ---
	userHandler := server.NewUserHandler(userService)
	transactionHandler := server.NewTransactionHandler(dispatcher, transactionService)
	authHandler := server.NewAuthHandler(userService, []byte(cfg.JWTSecret))
	balanceHandler := server.NewBalanceHandler(balanceService)

	srv := server.NewServer(cfg, log, userService, userHandler, transactionHandler, authHandler, balanceHandler)

	// --- Start Server and Handle Graceful Shutdown ---
	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, srv.Router()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("could not start server", "error", err)
			os.Exit(1)
		}
	}()

	fmt.Println("Press Ctrl+C to shut down.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("Shutdown signal received, shutting down gracefully...")
}
