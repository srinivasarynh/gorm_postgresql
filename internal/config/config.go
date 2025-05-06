package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	DB     DatabaseConfig
	Logger LoggerConfig
	App    AppConfig
}

type AppConfig struct {
	Environment string
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

type LoggerConfig struct {
	Level string
	Dev   bool
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	serverPort := getEnv("SERVER_PORT", "8000")
	readTimeout, _ := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "5"))
	writeTimeout, _ := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "10"))
	shutdownTimeout, _ := strconv.Atoi(getEnv("SERVER_SHUTDOWN_TIMEOUT", "5"))

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "app_db")
	dbSSLMode := getEnv("DB_SSL_MODE", "disable")
	dbMaxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	dbMaxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "25"))
	dbConnMaxLife, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "5"))

	logLevel := getEnv("LOG_LEVEL", "info")
	logDev, _ := strconv.ParseBool(getEnv("LOG_DEV", "false"))

	environment := getEnv("ENVIRONMENT", "development")

	return &Config{
		Server: ServerConfig{
			Port:            serverPort,
			ReadTimeout:     time.Duration(readTimeout) * time.Second,
			WriteTimeout:    time.Duration(writeTimeout) * time.Second,
			ShutdownTimeout: time.Duration(shutdownTimeout) * time.Second,
		},

		DB: DatabaseConfig{
			Host:         dbHost,
			Port:         dbPort,
			User:         dbUser,
			Password:     dbPassword,
			DBName:       dbName,
			SSLMode:      dbSSLMode,
			MaxOpenConns: dbMaxOpenConns,
			MaxIdleConns: dbMaxIdleConns,
			ConnMaxLife:  time.Duration(dbConnMaxLife) * time.Minute,
		},

		Logger: LoggerConfig{
			Level: logLevel,
			Dev:   logDev,
		},

		App: AppConfig{
			Environment: environment,
		},
	}, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
