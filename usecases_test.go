package betalinkauth_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	betalinkauth "github.com/BragdonD/betalink-auth"
	betalinklogger "github.com/BragdonD/betalink-logger"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dbContainer *postgres.PostgresContainer
	usecases    *betalinkauth.Usecases
	queries     *betalinkauth.Queries
	pgxConn     *pgx.Conn
	logger      *betalinklogger.Logger
)

const (
	dbUser     = "postgres"
	dbPassword = "postgres"
	dbName     = "auth"
)

const logPath = "./logs/betalink-auth-tests.log"

func TestMain(m *testing.M) {
	// Global setup
	ctx := context.Background()
	log.Println("Starting postgres container")
	var err error
	dbContainer, err = postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	defer func() {
		log.Println("Terminating postgres container")
		if err := testcontainers.TerminateContainer(dbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
		return
	}
	log.Println("Postgres container started: ", dbContainer.GetContainerID())

	endpoint, err := dbContainer.Endpoint(ctx, "")
	if err != nil {
		log.Fatalf("could not get postgres container endpoint: %v", err)
		return
	}
	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s/%s", dbUser, dbPassword, endpoint, dbName)
	pgxConn, err = pgx.Connect(ctx, postgresDSN)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
		return
	}
	defer pgxConn.Close(ctx)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(fmt.Errorf("could not open log file: %w", err))
	}
	defer logFile.Close()
	logger = betalinklogger.NewLogger("betalink-auth", false, true, logFile)

	// Run goose migrations
	migrationDir := "./migrations"
	if err := runGooseMigrations(migrationDir, postgresDSN); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func runGooseMigrations(migrationDir, dsn string) error {
	log.Println("Running goose migrations")
	cmd := exec.Command("/home/bragdon/.goose/bin/goose", "-dir", migrationDir, "postgres", dsn, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}
	log.Println("Migrations completed successfully")
	return nil
}

func TestNewUsecase(t *testing.T) {
	queries = betalinkauth.New(pgxConn)
	require.NotNil(t, *queries)
	log.Println("queries value: ", queries)
	usecases = betalinkauth.NewUsecase(logger, queries)
	require.NotNil(t, *usecases)
}

func TestUsecases_RegisterUser(t *testing.T) {
	if err := usecases.RegisterUser(
		context.Background(),
		"John", "Doe",
		"johndoe@gmail.com", "D.Ft[SHn5dLNb-wy=v'~$7"); err != nil {
		t.Fatalf("could not register user: %v", err)
		return
	}

	log.Println("User registered")

	// Test if the insertion was successful
	loginData, err := queries.GetLoginDataByEmail(context.Background(), "johndoe@gmail.com")
	if err != nil {
		t.Fatalf("could not get login data: %v", err)
		return
	}

	require.NotNil(t, loginData)
}
