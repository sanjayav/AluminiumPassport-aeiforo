
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "aluminium-passport/internal/models"
    "aluminium-passport/internal/services"
    "github.com/golang-jwt/jwt/v5"
    "strings"
    "os"
)

var users = map[string]struct {
    Password string
    Role     string
}{
    "admin":   {"admin123", models.RoleIssuer},
    "auditor": {"audit123", models.RoleAuditor},
    "viewer":  {"view123", models.RoleViewer},
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var creds models.Credentials
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    user, exists := users[creds.Username]
    if !exists || user.Password != creds.Password {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    token, err := services.GenerateToken(creds.Username, user.Role)
    if err != nil {
        http.Error(w, "Token generation failed", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func CreatePassportHandler(w http.ResponseWriter, r *http.Request) {
    var passport models.AluminiumPassport
    if err := json.NewDecoder(r.Body).Decode(&passport); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    if err := services.SavePassport(&passport); err != nil {
        http.Error(w, "Failed to save passport", http.StatusInternalServerError)
        return
    }

    user, role := extractUserRole(r)
    services.LogEvent(user, role, "CREATE", passport.PassportID)

    json.NewEncoder(w).Encode(map[string]string{"message": "Passport created"})
}

func GetPassportByIdHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    passport, err := services.GetPassportById(id)
    if err != nil {
        http.Error(w, "Passport not found", http.StatusNotFound)
        return
    }

    user, role := extractUserRole(r)
    services.LogEvent(user, role, "READ", id)

    json.NewEncoder(w).Encode(passport)
}

func extractUserRole(r *http.Request) (string, string) {
    tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
    token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })

    if claims, ok := token.Claims.(jwt.MapClaims); ok {
        user := claims["username"].(string)
        role := claims["role"].(string)
        return user, role
    }

    return "unknown", "unknown"
}
