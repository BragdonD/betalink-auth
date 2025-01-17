package betalinkauth

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"text/template"
)

// Config is the configuration for the auth service
type Config struct {
	// EnvironmentVarFile is the file containing the environment variables
	// If not provided, or an empty string, no environment variables will be loaded
	EnvironmentVarFile string `yaml:"env_file"`
	// Http is the configuration for the HTTP server
	Http HttpConfig `yaml:"http"`
	// DBConnTemplate is the template for the database connection string
	DBConnTemplate string `yaml:"db_conn_template"`
	// Auth is the configuration for the auth mechanism
	Auth AuthConfig `yaml:"auth"`
}

// ServerConfig is the configuration for the server
type HttpConfig struct {
	// Host is the server host
	Host string `yaml:"host"`
	// Port is the server port
	Port int `yaml:"port"`
	// TODO: add https support
}

// AuthConfig is the configuration for the auth mechanism
type AuthConfig struct {
	// AccessTokenConfig is the configuration for the access token
	AccessTokenConfig TokenConfig `yaml:"access_token"`
	// RefreshTokenConfig is the configuration for the refresh token
	RefreshTokenConfig TokenConfig `yaml:"refresh_token"`
}

// TokenConfig is the configuration for a JWT
// The secret can be provided in the yaml configuration file
// for the sake of simplicity. However, it should not be used
// in production. Instead use environment variables.
type TokenConfig struct {
	// Secret is the secret used to sign the token
	Secret string `yaml:"secret"`
	// Duration is the duration in seconds for which the token is valid
	Duration int `yaml:"duration"`
}

// DBConfig is the configuration for the database
// It can be provided in the yaml configuration file
// for the sake of simplicity. However, it should not
// be used in production. Instead use environment variables.
type DBConfig struct {
	// Host is the database host
	Host string `yaml:"host"`
	// Port is the database port
	Port int `yaml:"port"`
	// User is the database user
	User string `yaml:"user"`
	// Password is the database password
	Password string `yaml:"password"`
	// Name is the database name
	Name string `yaml:"name"`
}

// LoadDBConfigFromEnv loads the database configuration
// from environment variables
func LoadDBConfigFromEnv() (*DBConfig, error) {
	portStr, ok := os.LookupEnv("DB_PORT")
	if !ok {
		portStr = "5432"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}
	host, ok := os.LookupEnv("DB_HOST")
	if !ok {
		host = "localhost"
	}
	user, ok := os.LookupEnv("DB_USER")
	if !ok {
		user = "betalinkauth"
	}
	password, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		password = "betalinkauth"
	}
	name, ok := os.LookupEnv("DB_NAME")
	if !ok {
		name = "betalinkauth"
	}

	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
	}, nil
}

// GetDBConnString returns the connection string for the database
// based on the template string and the configuration
func (c *DBConfig) GetDBConnString(conn_template string) (string, error) {
	templ, err := template.New("db_conn").Parse(conn_template)
	if err != nil {
		return "", fmt.Errorf("could not parse template: %w", err)
	}

	var buf bytes.Buffer
	err = templ.Execute(&buf, map[string]interface{}{
		"user":     c.User,
		"password": c.Password,
		"host":     c.Host,
		"port":     c.Port,
		"dbname":   c.Name,
	})
	if err != nil {
		return "", fmt.Errorf("could not execute template: %w", err)
	}

	return buf.String(), nil
}
