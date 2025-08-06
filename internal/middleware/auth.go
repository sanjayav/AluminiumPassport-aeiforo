package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"aluminium-passport/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			if err == auth.ErrExpiredToken {
				http.Error(w, "Token has expired", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddlewareFunc wraps a handler function with authentication
func AuthMiddlewareFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(handlerFunc).ServeHTTP
}

// RoleMiddleware checks if user has required role(s)
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (should be set by AuthMiddleware)
			claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
			if !ok {
				http.Error(w, "User context not found", http.StatusInternalServerError)
				return
			}

			// Check if user role is in allowed roles
			authorized := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					authorized = true
					break
				}
			}

			if !authorized {
				http.Error(w, fmt.Sprintf("Insufficient permissions. Required roles: %v", allowedRoles), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddlewareFunc wraps a handler function with role checking
func RoleMiddlewareFunc(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return RoleMiddleware(allowedRoles...)(handlerFunc).ServeHTTP
	}
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(r *http.Request) (*auth.Claims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
	return claims, ok
}

// RequireAuth is a convenience function that combines auth and role checking
func RequireAuth(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return AuthMiddleware(RoleMiddleware(allowedRoles...)(next))
	}
}

// OptionalAuth middleware that validates token if present but doesn't require it
func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				if claims, err := auth.ValidateToken(tokenString); err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, claims)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
