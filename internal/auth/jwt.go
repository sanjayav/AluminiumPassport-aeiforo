package auth

import (
	"errors"
	"fmt"
	"time"

	"aluminium-passport/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	WalletAddr  string `json:"wallet_address"`
	CompanyName string `json:"company_name"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(userID int, username, email, role, walletAddr, companyName string) (*TokenPair, error) {
	cfg := config.AppConfig

	// Access token (shorter expiration)
	accessExpirationTime := time.Now().Add(time.Duration(cfg.JWTExpirationHours) * time.Hour)
	accessClaims := &Claims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Role:        role,
		WalletAddr:  walletAddr,
		CompanyName: companyName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aluminium-passport-api",
			Subject:   fmt.Sprintf("user:%d", userID),
			ID:        fmt.Sprintf("%d-%d", userID, time.Now().Unix()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Refresh token (longer expiration)
	refreshExpirationTime := time.Now().Add(time.Duration(cfg.JWTRefreshHours) * time.Hour)
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aluminium-passport-api",
			Subject:   fmt.Sprintf("refresh:%d", userID),
			ID:        fmt.Sprintf("refresh-%d-%d", userID, time.Now().Unix()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    accessExpirationTime.Unix(),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfig

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshToken generates a new access token from a valid refresh token
func RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Check if this is actually a refresh token
	if claims.Subject[:7] != "refresh" {
		return nil, ErrInvalidToken
	}

	// Generate new token pair
	return GenerateTokenPair(
		claims.UserID,
		claims.Username,
		claims.Email,
		claims.Role,
		claims.WalletAddr,
		claims.CompanyName,
	)
}

// ExtractClaims extracts claims from token without validation (for expired tokens)
func ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(tokenString string) bool {
	claims, err := ExtractClaims(tokenString)
	if err != nil {
		return true
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}

// GetUserRoleFromToken extracts user role from token
func GetUserRoleFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.Role, nil
}

// GetUserIDFromToken extracts user ID from token
func GetUserIDFromToken(tokenString string) (int, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}
