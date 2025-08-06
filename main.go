package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aluminium-passport/internal/config"
	"aluminium-passport/internal/db"
	"aluminium-passport/internal/ipfs"
	"aluminium-passport/internal/routes"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := db.InitializeDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Initialize IPFS client
	if err := ipfs.InitializeIPFS(); err != nil {
		log.Printf("Warning: Failed to initialize IPFS client: %v", err)
		log.Println("IPFS functionality will be disabled")
	} else {
		log.Println("IPFS client initialized successfully")
	}

	// Setup routes
	router := routes.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Aluminium Passport API server starting on port %s", cfg.Port)
		log.Printf("📊 Environment: %s", cfg.Environment)
		log.Printf("🔗 Database: Connected")

		if ipfs.IsIPFSAvailable() {
			log.Printf("🌐 IPFS: Connected")
		} else {
			log.Printf("🌐 IPFS: Disconnected")
		}

		log.Printf("🔐 JWT: Configured")
		log.Printf("📝 API Documentation: http://localhost:%s/health", cfg.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server shutdown completed")
	}

	// Close database connection
	if err := db.CloseDB(); err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		log.Println("✅ Database connection closed")
	}

	log.Println("🎉 Aluminium Passport API stopped")
}
