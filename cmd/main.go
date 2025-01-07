package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	betalinkauth "github.com/BragdonD/betalink-auth"
	betalinklogger "github.com/BragdonD/betalink-logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// TODO: implement configuration
const (
	logPath = "./logs/betalink-auth.log"
)

func main() {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(fmt.Errorf("could not open log file: %w", err))
	}
	defer logFile.Close()
	logger := betalinklogger.NewLogger("betalink-auth", false, true, logFile)
	logger.Info("Starting betalink-auth service")

	logger.Info("Opening database connection")
	conn, err := pgx.Connect(context.Background(), "postgres://betalinkauth:betalinkauth@localhost:5432/betalinkauth")
	if err != nil {
		logger.Error(fmt.Errorf("could not connect to database: %w", err))
		return
	}
	defer conn.Close(context.Background())

	logger.Info("Initializing http server")
	queries := betalinkauth.New(conn)
	usecase := betalinkauth.NewUsecase(logger, queries)

	ginRouter := gin.Default()
	betalinkauth.NewRouter(logger, ginRouter, usecase)

	// Create the HTTP server with Gin's router
	srv := &http.Server{
		Addr:    ":8080",
		Handler: ginRouter,
	}

	// Channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not start server: %v\n", err)
		}
	}()

	// Wait for OS signal
	<-quit
	logger.Info("Shutting down server...")

	// Context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v\n", err)
	}

	logger.Info("Server exiting")
}
