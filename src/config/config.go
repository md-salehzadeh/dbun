package config

import (
	"os"
	"strconv"
)

// DBConfig holds database connection settings
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// LoadConfig loads database configuration from environment variables with fallbacks
func LoadConfig() DBConfig {
	// Default database config
	config := DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "",
		Database: "test",
	}

	// Override with environment variables if available
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Host = host
	}

	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	if user := os.Getenv("DB_USER"); user != "" {
		config.User = user
	}

	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Password = password
	}

	if database := os.Getenv("DB_NAME"); database != "" {
		config.Database = database
	}

	return config
}
