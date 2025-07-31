
package middleware

import (
    "fmt"
    "net/http"
    "os"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

func RoleGuard(requiredRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
            if tokenStr == "" {
                http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
                return
            }

            token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
                return []byte(os.Getenv("JWT_SECRET")), nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                http.Error(w, "Unauthorized: Invalid claims", http.StatusUnauthorized)
                return
            }

            role, ok := claims["role"].(string)
            if !ok {
                http.Error(w, "Unauthorized: Missing role", http.StatusUnauthorized)
                return
            }

            authorized := false
            for _, r := range requiredRoles {
                if role == r {
                    authorized = true
                    break
                }
            }

            if !authorized {
                http.Error(w, fmt.Sprintf("Forbidden: Role %s not allowed", role), http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
