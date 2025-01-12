package betalinkauth_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	betalinkauth "github.com/BragdonD/betalink-auth"
	betalinklogger "github.com/BragdonD/betalink-logger"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testCtx     context.Context
	dbContainer *postgres.PostgresContainer
)

const (
	dbUser     = "auth"
	dbPassword = "auth"
	dbName     = "auth"
)

const logPath = "./logs/betalink-auth-tests.log"

func TestMain(m *testing.M) {
	// Global setup
	testCtx = context.Background()
	log.Println("Starting postgres container")
	var err error
	dbContainer, err = postgres.Run(testCtx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
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

	// Run goose migrations
	dbURL, err := dbContainer.ConnectionString(testCtx)
	if err != nil {
		log.Fatalf("could not get connection string: %v", err)
		return
	}
	migrationDir := "./migrations"
	if err := runGooseMigrations(migrationDir, dbURL); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("Creating snapshot of the container")
	err = dbContainer.Snapshot(testCtx)
	if err != nil {
		log.Fatalf("could not create snapshot of the container: %v", err)
		return
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

func createPgxConn() (*pgx.Conn, error) {
	dbURL, err := dbContainer.ConnectionString(testCtx)
	if err != nil {
		return nil, fmt.Errorf("could not get connection string: %w", err)
	}

	conn, err := pgx.Connect(testCtx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return conn, nil
}

func createLogger() (*betalinklogger.Logger, error) {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %w", err)
	}
	return betalinklogger.NewLogger("betalink-auth", false, true, logFile), nil
}

func TestNewUsecase(t *testing.T) {
	conn, err := createPgxConn()
	if err != nil {
		t.Fatalf("could not create pgx connection: %v", err)
	}
	defer conn.Close(context.Background())
	queries := betalinkauth.New(conn)
	logger, err := createLogger()
	if err != nil {
		t.Fatalf("could not create logger: %v", err)
	}
	usecases := betalinkauth.NewUsecase(logger, queries)
	require.NotNil(t, usecases)
}

func TestUsecases_RegisterUser(t *testing.T) {
	err := dbContainer.Restore(testCtx)
	require.NoError(t, err)

	conn, err := createPgxConn()
	if err != nil {
		t.Fatalf("could not create pgx connection: %v", err)
		return
	}
	defer conn.Close(context.Background())

	queries := betalinkauth.New(conn)
	logger, err := createLogger()
	if err != nil {
		t.Fatalf("could not create logger: %v", err)
	}

	usecases := betalinkauth.NewUsecase(logger, queries)

	if err := usecases.RegisterUser(
		context.Background(),
		"John", "Doe",
		"johndoe@gmail.com", "D.Ft[SHn5dLNb-wy=v'~$7"); err != nil {
		t.Fatalf("could not register user: %v", err)
		return
	}

	log.Println("User registered")

	// Test if the insertion was successful
	loginData, err := queries.GetLoginDataByEmail(testCtx, "johndoe@gmail.com")
	if err != nil {
		t.Fatalf("could not get login data: %v", err)
		return
	}

	require.NotNil(t, loginData)
}

func TestUsecases_LoginUser(t *testing.T) {
	err := dbContainer.Restore(testCtx)
	require.NoError(t, err)

	conn, err := createPgxConn()
	if err != nil {
		t.Fatalf("could not create pgx connection: %v", err)
		return
	}
	defer conn.Close(context.Background())

	queries := betalinkauth.New(conn)
	logger, err := createLogger()
	if err != nil {
		t.Fatalf("could not create logger: %v", err)
	}

	usecases := betalinkauth.NewUsecase(logger, queries)

	if err := usecases.RegisterUser(
		context.Background(),
		"John", "Doe",
		"johndoe@gmail.com", "D.Ft[SHn5dLNb-wy=v'~$7"); err != nil {
		t.Fatalf("could not register user: %v", err)
		return
	}

	log.Println("User registered")

	if _, err := usecases.LoginUser(testCtx,
		"johndoe@gmail.com", "D.Ft[SHn5dLNb-wy=v'~$7"); err != nil {
		t.Fatalf("could not login user: %v", err)
		return
	}

	if _, err := usecases.LoginUser(testCtx,
		"johndoe@gmail.com", "testPassword"); err == nil {
		t.Fatalf("user can login with wrong password: %v", err)
		return
	}
}
