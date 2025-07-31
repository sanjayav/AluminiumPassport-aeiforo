
package services

import (
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

func GenerateToken(username, role string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := jwt.MapClaims{
        "username": username,
        "role":     role,
        "exp":      expirationTime.Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
