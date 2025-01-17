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
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// TODO: implement configuration
const (
	logPath = "./logs/betalink-auth.log"
)

// loadYamlConfig loads the configuration from a yaml file
func loadYamlConfig(path string) (*betalinkauth.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	config := &betalinkauth.Config{}
	if err := yaml.NewDecoder(file).Decode(config); err != nil {
		return nil, fmt.Errorf("could not decode config file: %w", err)
	}

	return config, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./main <config-path>")
		os.Exit(1)
	}
	configPath := os.Args[1]

	// Load the configuration
	config, err := loadYamlConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("could not load config: %w", err))
	}
	if config.EnvironmentVarFile != "" {
		if err := godotenv.Load(config.EnvironmentVarFile); err != nil {
			panic(fmt.Errorf("could not load environment variables: %w", err))
		}
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(fmt.Errorf("could not open log file: %w", err))
	}
	defer logFile.Close()
	logger := betalinklogger.NewLogger("betalink-auth", false, true, logFile)
	logger.Info("Starting betalink-auth service")

	dbConfig, err := betalinkauth.LoadDBConfigFromEnv()
	if err != nil {
		logger.Error(fmt.Errorf("could not load db config: %w", err))
		return
	}

	connStr, err := dbConfig.GetDBConnString(config.DBConnTemplate)
	if err != nil {
		logger.Error(fmt.Errorf("could not get db connection string: %w", err))
		return
	}

	logger.Info("Opening database connection")
	conn, err := pgx.Connect(context.Background(), connStr)
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
