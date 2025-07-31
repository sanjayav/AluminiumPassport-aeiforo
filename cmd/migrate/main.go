
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
)

func main() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL not set")
    }

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Failed to connect to DB:", err)
    }
    defer db.Close()

    sqlBytes, err := os.ReadFile("db/migrations/init.sql")
    if err != nil {
        log.Fatal("Failed to read SQL file:", err)
    }

    _, err = db.Exec(string(sqlBytes))
    if err != nil {
        log.Fatal("Migration failed:", err)
    }

    fmt.Println("Migration executed successfully.")
}
