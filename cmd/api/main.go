package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go_postgres/internal/config"
	"go_postgres/internal/db"
	"go_postgres/internal/db/migrations"
	"go_postgres/internal/handlers"
	"go_postgres/internal/middleware"
	"go_postgres/internal/repository"
	"go_postgres/internal/service"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(cfg.Logger)
	defer logger.Sync()

	// Run database migrations
	logger.Info("Running database migrations...")
	if err := migrations.RunMigrations(cfg.DB.GetDSN()); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Connect to the database
	db, err := db.NewPostgresDB(&cfg.DB, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB, logger)

	// Initialize services
	userService := service.NewUserService(userRepo, logger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger)

	// Set up routes
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /api/auth/login", userHandler.AuthenticateUser)
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)

	// Protected routes
	authRouter := middleware.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/users":
			userHandler.ListUsers(w, r)
		case r.Method == http.MethodGet && r.PathValue("id") != "":
			userHandler.GetUser(w, r)
		case r.Method == http.MethodPut && r.PathValue("id") != "":
			userHandler.UpdateUser(w, r)
		case r.Method == http.MethodDelete && r.PathValue("id") != "":
			userHandler.DeleteUser(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	mux.Handle("GET /api/users", authRouter)
	mux.Handle("GET /api/users/{id}", authRouter)
	mux.Handle("PUT /api/users/{id}", authRouter)
	mux.Handle("DELETE /api/users/{id}", authRouter)

	// Set up middleware
	handler := middleware.RequestLogger(logger)(mux)

	// Initialize server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}

// initLogger initializes the logger
func initLogger(cfg config.LoggerConfig) *zap.Logger {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	var logger *zap.Logger
	var err error

	if cfg.Dev {
		// Development logger
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(level)
		logger, err = config.Build()
	} else {
		// Production logger
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(level)
		logger, err = config.Build()
	}

	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	return logger
}
