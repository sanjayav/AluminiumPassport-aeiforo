
#!/bin/sh

echo "Starting Aluminium Passport Backend Docker container..."

# Print environment
echo "Running with ENV:"
env

# Run migration if needed
go run cmd/migrate/main.go

# Start the backend
go run main.go
