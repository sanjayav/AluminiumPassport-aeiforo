package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aluminium-passport/internal/auth"
	"aluminium-passport/internal/db"
	"aluminium-passport/internal/models"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Username      string `json:"username" binding:"required"`
	Email         string `json:"email"`
	Password      string `json:"password" binding:"required"`
	Role          string `json:"role"`
	CompanyName   string `json:"company_name"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	User    *UserResponse   `json:"user"`
	Tokens  *auth.TokenPair `json:"tokens"`
	Message string          `json:"message"`
}

type UserResponse struct {
	ID            int         `json:"id"`
	WalletAddress string      `json:"wallet_address"`
	Username      string      `json:"username"`
	Email         *string     `json:"email"`
	Role          string      `json:"role"`
	CompanyName   *string     `json:"company_name"`
	ContactInfo   *db.JSONMap `json:"contact_info"`
	IsActive      bool        `json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	LastLogin     *time.Time  `json:"last_login"`
}

// Login authenticates a user and returns JWT tokens
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := ac.getUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is deactivated", http.StatusForbidden)
		return
	}

	// Verify password
	if err := auth.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token pair
	tokens, err := auth.GenerateTokenPair(
		user.ID,
		user.Username,
		getStringValue(user.Email),
		user.Role,
		user.WalletAddress,
		getStringValue(user.CompanyName),
	)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Update last login
	if err := ac.updateLastLogin(user.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login for user %d: %v\n", user.ID, err)
	}

	// Log audit event
	ac.logAuditEvent(user.ID, user.Role, "LOGIN", "user", fmt.Sprintf("%d", user.ID), nil, nil, r)

	// Prepare response
	userResponse := &UserResponse{
		ID:            user.ID,
		WalletAddress: user.WalletAddress,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		CompanyName:   user.CompanyName,
		ContactInfo:   user.ContactInfo,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		LastLogin:     user.LastLogin,
	}

	response := &AuthResponse{
		User:    userResponse,
		Tokens:  tokens,
		Message: "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Register creates a new user account
func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate wallet address format
	if !isValidWalletAddress(req.WalletAddress) {
		http.Error(w, "Invalid wallet address format", http.StatusBadRequest)
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = models.RoleViewer
	}

	// Validate role
	validRoles := models.GetValidRoles()
	if !contains(validRoles, req.Role) {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Check if role requires super admin approval (suppliers)
	if models.RequiresSuperAdminApproval(req.Role) {
		http.Error(w, "Supplier roles require admin approval through the onboarding process", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := auth.ValidateAndHashPassword(req.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Password validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	if exists, err := ac.userExists(req.Username, req.WalletAddress); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Create user
	user := &db.User{
		WalletAddress: req.WalletAddress,
		Username:      req.Username,
		Email:         nullableString(req.Email),
		PasswordHash:  hashedPassword,
		Role:          req.Role,
		CompanyName:   nullableString(req.CompanyName),
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	userID, err := ac.createUser(user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.ID = userID

	// Generate token pair
	tokens, err := auth.GenerateTokenPair(
		user.ID,
		user.Username,
		getStringValue(user.Email),
		user.Role,
		user.WalletAddress,
		getStringValue(user.CompanyName),
	)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Log audit event
	ac.logAuditEvent(user.ID, user.Role, "REGISTER", "user", fmt.Sprintf("%d", user.ID), nil, nil, r)

	// Prepare response
	userResponse := &UserResponse{
		ID:            user.ID,
		WalletAddress: user.WalletAddress,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		CompanyName:   user.CompanyName,
		ContactInfo:   user.ContactInfo,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		LastLogin:     user.LastLogin,
	}

	response := &AuthResponse{
		User:    userResponse,
		Tokens:  tokens,
		Message: "Registration successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken generates new tokens from a refresh token
func (ac *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Refresh tokens
	tokens, err := auth.RefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Get user info from token
	claims, err := auth.ValidateToken(tokens.AccessToken)
	if err != nil {
		http.Error(w, "Failed to validate new token", http.StatusInternalServerError)
		return
	}

	// Log audit event
	ac.logAuditEvent(claims.UserID, claims.Role, "REFRESH_TOKEN", "token", "refresh", nil, nil, r)

	response := map[string]interface{}{
		"tokens":  tokens,
		"message": "Token refreshed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout invalidates the user's session (placeholder - would need token blacklisting in production)
func (ac *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Log audit event
	ac.logAuditEvent(claims.UserID, claims.Role, "LOGOUT", "user", fmt.Sprintf("%d", claims.UserID), nil, nil, r)

	// In production, you would add the token to a blacklist here

	response := map[string]string{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile returns the current user's profile
func (ac *AuthController) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := ac.getUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare response
	userResponse := &UserResponse{
		ID:            user.ID,
		WalletAddress: user.WalletAddress,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		CompanyName:   user.CompanyName,
		ContactInfo:   user.ContactInfo,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		LastLogin:     user.LastLogin,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponse)
}

// Database helper methods
func (ac *AuthController) getUserByUsername(username string) (*db.User, error) {
	query := `
		SELECT id, wallet_address, username, email, password_hash, role, 
		       company_name, contact_info, is_active, created_at, updated_at, last_login
		FROM users 
		WHERE username = $1`

	user := &db.User{}
	err := db.DB.QueryRow(query, username).Scan(
		&user.ID, &user.WalletAddress, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CompanyName, &user.ContactInfo, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)

	return user, err
}

func (ac *AuthController) getUserByID(userID int) (*db.User, error) {
	query := `
		SELECT id, wallet_address, username, email, password_hash, role, 
		       company_name, contact_info, is_active, created_at, updated_at, last_login
		FROM users 
		WHERE id = $1`

	user := &db.User{}
	err := db.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.WalletAddress, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CompanyName, &user.ContactInfo, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)

	return user, err
}

func (ac *AuthController) userExists(username, walletAddress string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR wallet_address = $2)`

	var exists bool
	err := db.DB.QueryRow(query, username, walletAddress).Scan(&exists)
	return exists, err
}

func (ac *AuthController) createUser(user *db.User) (int, error) {
	query := `
		INSERT INTO users (wallet_address, username, email, password_hash, role, company_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var userID int
	err := db.DB.QueryRow(
		query,
		user.WalletAddress, user.Username, user.Email, user.PasswordHash,
		user.Role, user.CompanyName, user.IsActive, user.CreatedAt, user.UpdatedAt,
	).Scan(&userID)

	return userID, err
}

func (ac *AuthController) updateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = $1 WHERE id = $2`
	_, err := db.DB.Exec(query, time.Now(), userID)
	return err
}

func (ac *AuthController) logAuditEvent(userID int, userRole, action, resourceType, resourceID string, oldValues, newValues *db.JSONMap, r *http.Request) {
	// This would log to the audit_logs table
	// Implementation depends on your audit logging requirements
}

// Helper functions
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func isValidWalletAddress(address string) bool {
	// Basic Ethereum address validation
	return len(address) == 42 && strings.HasPrefix(address, "0x")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
