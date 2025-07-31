
package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"
    "github.com/rs/cors"
    "net/http"

    "aluminium-passport/internal/router"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, continuing with environment variables")
    }

    // Initialize the router
    r := router.SetupRouter()

    // Add CORS middleware
    handler := cors.Default().Handler(r)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Aluminium Passport backend running on http://localhost:%s", port)
    if err := http.ListenAndServe(":"+port, handler); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
