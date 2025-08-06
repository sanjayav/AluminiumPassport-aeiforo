package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"aluminium-passport/internal/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitializeDB initializes the database connection
func InitializeDB() error {
	cfg := config.AppConfig

	var err error
	DB, err = sql.Open("postgres", cfg.GetDatabaseDSN())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func GetStats() sql.DBStats {
	if DB == nil {
		return sql.DBStats{}
	}
	return DB.Stats()
}
