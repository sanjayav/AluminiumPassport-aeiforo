package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aluminium-passport/internal/auth"
	"aluminium-passport/internal/db"
	"aluminium-passport/internal/models"

	"github.com/gorilla/mux"
)

type ApprovalController struct{}

func NewApprovalController() *ApprovalController {
	return &ApprovalController{}
}

// CreateApprovalRequest creates a new approval request
func (ac *ApprovalController) CreateApprovalRequest(w http.ResponseWriter, r *http.Request) {
	claims, err := ac.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ApprovalRequestCreateData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request type and approver role
	if !ac.isValidApprovalType(string(req.RequestType)) {
		http.Error(w, "Invalid request type", http.StatusBadRequest)
		return
	}

	if !ac.isValidRole(req.ApproverRole) {
		http.Error(w, "Invalid approver role", http.StatusBadRequest)
		return
	}

	// Set default expiry if not provided
	expiresInDays := req.ExpiresInDays
	if expiresInDays <= 0 {
		expiresInDays = 7 // Default 7 days
	}

	// Create approval request
	approvalRequest := &models.ApprovalRequest{
		RequestType:  req.RequestType,
		RequestedBy:  claims.UserID,
		ApproverRole: req.ApproverRole,
		Status:       models.ApprovalStatusPending,
		Title:        req.Title,
		Description:  req.Description,
		RequestData:  req.RequestData,
		ExpiresAt:    timePtr(time.Now().AddDate(0, 0, expiresInDays)),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	requestID, err := ac.createApprovalRequest(approvalRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create approval request: %v", err), http.StatusInternalServerError)
		return
	}

	approvalRequest.ID = requestID

	// Log audit event
	ac.logAuditEvent(claims.UserID, claims.Role, "CREATE", "approval_request", fmt.Sprintf("%d", requestID), nil, approvalRequest, r)

	// Send notification to approvers (placeholder)
	ac.notifyApprovers(approvalRequest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Approval request created successfully",
		"request_id": requestID,
		"status":     "pending",
		"expires_at": approvalRequest.ExpiresAt,
	})
}

// RequestSupplierOnboarding creates an approval request for supplier onboarding
func (ac *ApprovalController) RequestSupplierOnboarding(w http.ResponseWriter, r *http.Request) {
	claims, err := ac.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only admins can request supplier onboarding
	if !models.HasHigherOrEqualRole(claims.Role, models.RoleAdmin) {
		http.Error(w, "Only admins can request supplier onboarding", http.StatusForbidden)
		return
	}

	var req models.SupplierOnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the requested role is a supplier role
	if !models.RequiresSuperAdminApproval(req.Role) {
		http.Error(w, "Requested role does not require super admin approval", http.StatusBadRequest)
		return
	}

	// Validate wallet address format
	if !isValidWalletAddress(req.WalletAddress) {
		http.Error(w, "Invalid wallet address format", http.StatusBadRequest)
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

	// Hash the password (assuming it's provided in the request)
	hashedPassword, err := auth.ValidateAndHashPassword("TempPassword123!") // Temporary password
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create pending user record
	pendingUser := &models.PendingUser{
		WalletAddress:   req.WalletAddress,
		Username:        req.Username,
		Email:           nullableString(req.Email),
		PasswordHash:    hashedPassword,
		RequestedRole:   req.Role,
		CompanyName:     nullableString(req.CompanyName),
		CompanyType:     nullableString(req.CompanyType),
		BusinessLicense: nullableString(req.BusinessLicense),
		ContactInfo:     &req.ContactInfo,
		Justification:   nullableString(req.Justification),
		CreatedAt:       time.Now(),
	}

	// Create approval request data
	requestData := &models.JSONMap{
		"supplier_data": req,
	}

	approvalRequest := &models.ApprovalRequest{
		RequestType:  models.ApprovalTypeSupplierOnboarding,
		RequestedBy:  claims.UserID,
		ApproverRole: models.RoleSuperAdmin,
		Status:       models.ApprovalStatusPending,
		Title:        fmt.Sprintf("Supplier Onboarding: %s (%s)", req.CompanyName, req.Role),
		Description:  fmt.Sprintf("Admin %s requests onboarding of %s as %s. Justification: %s", claims.Username, req.CompanyName, req.Role, req.Justification),
		RequestData:  requestData,
		ExpiresAt:    timePtr(time.Now().AddDate(0, 0, 7)), // 7 days expiry
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Begin transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Database transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create approval request
	requestID, err := ac.createApprovalRequestTx(tx, approvalRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create approval request: %v", err), http.StatusInternalServerError)
		return
	}

	// Create pending user with reference to approval request
	pendingUser.ApprovalRequestID = requestID
	if err := ac.createPendingUserTx(tx, pendingUser); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create pending user: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Log audit event
	ac.logAuditEvent(claims.UserID, claims.Role, "CREATE", "supplier_onboarding_request", fmt.Sprintf("%d", requestID), nil, req, r)

	// Send notification to super admins
	ac.notifyApprovers(approvalRequest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":             "Supplier onboarding request created successfully",
		"approval_request_id": requestID,
		"status":              "pending_super_admin_approval",
		"expires_at":          approvalRequest.ExpiresAt,
		"approver_role":       models.RoleSuperAdmin,
	})
}

// GetApprovalRequests returns paginated list of approval requests
func (ac *ApprovalController) GetApprovalRequests(w http.ResponseWriter, r *http.Request) {
	claims, err := ac.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status := r.URL.Query().Get("status")
	requestType := r.URL.Query().Get("type")
	forApproval := r.URL.Query().Get("for_approval") == "true"

	// Get approval requests
	requests, total, err := ac.listApprovalRequests(page, limit, status, requestType, forApproval, claims.Role, claims.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve approval requests", http.StatusInternalServerError)
		return
	}

	// Log audit event
	ac.logAuditEvent(claims.UserID, claims.Role, "VIEW", "approval_requests", "", nil, nil, r)

	response := map[string]interface{}{
		"requests":    requests,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetApprovalRequest returns a specific approval request
func (ac *ApprovalController) GetApprovalRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	claims, err := ac.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get approval request
	approvalRequest, err := ac.getApprovalRequestByID(requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Approval request not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user can view this request
	if !ac.canViewApprovalRequest(approvalRequest, claims.Role, claims.UserID) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Log audit event
	ac.logAuditEvent(claims.UserID, claims.Role, "VIEW", "approval_request", fmt.Sprintf("%d", requestID), nil, nil, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(approvalRequest)
}

// ApproveRequest approves an approval request
func (ac *ApprovalController) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	ac.processApprovalAction(w, r, models.ApprovalStatusApproved)
}

// RejectRequest rejects an approval request
func (ac *ApprovalController) RejectRequest(w http.ResponseWriter, r *http.Request) {
	ac.processApprovalAction(w, r, models.ApprovalStatusRejected)
}

// processApprovalAction handles approval or rejection of requests
func (ac *ApprovalController) processApprovalAction(w http.ResponseWriter, r *http.Request, action models.ApprovalStatus) {
	vars := mux.Vars(r)
	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	claims, err := ac.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var actionData models.ApprovalAction
	if err := json.NewDecoder(r.Body).Decode(&actionData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get approval request
	approvalRequest, err := ac.getApprovalRequestByID(requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Approval request not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user can approve this request
	if !approvalRequest.CanBeApprovedBy(claims.Role) {
		http.Error(w, "Insufficient permissions to approve this request", http.StatusForbidden)
		return
	}

	// Check if request is still pending
	if !approvalRequest.IsPending() {
		http.Error(w, "Request is not pending or has expired", http.StatusBadRequest)
		return
	}

	// Begin transaction for approval process
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Database transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update approval request
	now := time.Now()
	updateData := map[string]interface{}{
		"status":      action,
		"approved_by": claims.UserID,
		"updated_at":  now,
		"approved_at": now,
	}

	if action == models.ApprovalStatusApproved {
		updateData["approval_reason"] = actionData.Reason
	} else {
		updateData["rejection_reason"] = actionData.Reason
	}

	if err := ac.updateApprovalRequestTx(tx, requestID, updateData); err != nil {
		http.Error(w, "Failed to update approval request", http.StatusInternalServerError)
		return
	}

	// Process the approval based on request type
	if action == models.ApprovalStatusApproved {
		if err := ac.processApprovedRequest(tx, approvalRequest); err != nil {
			http.Error(w, fmt.Sprintf("Failed to process approved request: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Log audit event
	auditAction := "APPROVE"
	if action == models.ApprovalStatusRejected {
		auditAction = "REJECT"
	}
	ac.logAuditEvent(claims.UserID, claims.Role, auditAction, "approval_request", fmt.Sprintf("%d", requestID), approvalRequest, updateData, r)

	// Send notification to requester
	ac.notifyRequester(approvalRequest, action, actionData.Reason)

	actionStr := "approved"
	if action == models.ApprovalStatusRejected {
		actionStr = "rejected"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     fmt.Sprintf("Request %s successfully", actionStr),
		"status":      action,
		"reason":      actionData.Reason,
		"actioned_by": claims.Username,
		"actioned_at": now,
	})
}

// Database helper methods
func (ac *ApprovalController) createApprovalRequest(req *models.ApprovalRequest) (int, error) {
	query := `
		INSERT INTO approval_requests (request_type, requested_by, approver_role, status, title, description, request_data, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	var requestID int
	err := db.DB.QueryRow(
		query,
		req.RequestType, req.RequestedBy, req.ApproverRole, req.Status,
		req.Title, req.Description, req.RequestData, req.ExpiresAt,
		req.CreatedAt, req.UpdatedAt,
	).Scan(&requestID)

	return requestID, err
}

func (ac *ApprovalController) createApprovalRequestTx(tx *sql.Tx, req *models.ApprovalRequest) (int, error) {
	query := `
		INSERT INTO approval_requests (request_type, requested_by, approver_role, status, title, description, request_data, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	var requestID int
	err := tx.QueryRow(
		query,
		req.RequestType, req.RequestedBy, req.ApproverRole, req.Status,
		req.Title, req.Description, req.RequestData, req.ExpiresAt,
		req.CreatedAt, req.UpdatedAt,
	).Scan(&requestID)

	return requestID, err
}

// Additional database methods and helper functions would continue here...
// (Implementation continues with similar patterns for other database operations)

// Helper methods
func (ac *ApprovalController) extractUserClaims(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return auth.ValidateToken(tokenString)
}

func (ac *ApprovalController) isValidApprovalType(requestType string) bool {
	validTypes := []string{"supplier_onboarding", "user_role_change", "system_configuration"}
	for _, t := range validTypes {
		if t == requestType {
			return true
		}
	}
	return false
}

func (ac *ApprovalController) isValidRole(role string) bool {
	validRoles := models.GetValidRoles()
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (ac *ApprovalController) logAuditEvent(userID int, userRole, action, resourceType, resourceID string, oldValues, newValues interface{}, r *http.Request) {
	// Implementation for audit logging
}

func (ac *ApprovalController) notifyApprovers(req *models.ApprovalRequest) {
	// Implementation for sending notifications to approvers
}

func (ac *ApprovalController) notifyRequester(req *models.ApprovalRequest, action models.ApprovalStatus, reason string) {
	// Implementation for sending notifications to requester
}

// Placeholder implementations for remaining methods
func (ac *ApprovalController) userExists(username, walletAddress string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR wallet_address = $2)`
	var exists bool
	err := db.DB.QueryRow(query, username, walletAddress).Scan(&exists)
	return exists, err
}

func (ac *ApprovalController) getApprovalRequestByID(requestID int) (*models.ApprovalRequest, error) {
	// Implementation to fetch approval request by ID
	return nil, nil // Placeholder
}

func (ac *ApprovalController) canViewApprovalRequest(req *models.ApprovalRequest, userRole string, userID int) bool {
	// Implementation to check viewing permissions
	return true // Placeholder
}

func (ac *ApprovalController) listApprovalRequests(page, limit int, status, requestType string, forApproval bool, userRole string, userID int) ([]*models.ApprovalRequest, int, error) {
	// Implementation to list approval requests with filtering
	return nil, 0, nil // Placeholder
}

func (ac *ApprovalController) updateApprovalRequestTx(tx *sql.Tx, requestID int, updateData map[string]interface{}) error {
	// Implementation to update approval request in transaction
	return nil // Placeholder
}

func (ac *ApprovalController) processApprovedRequest(tx *sql.Tx, req *models.ApprovalRequest) error {
	// Implementation to process approved requests (e.g., create user for supplier onboarding)
	return nil // Placeholder
}

func (ac *ApprovalController) createPendingUserTx(tx *sql.Tx, user *models.PendingUser) error {
	// Implementation to create pending user record
	return nil // Placeholder
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isValidWalletAddress(address string) bool {
	return len(address) == 42 && strings.HasPrefix(address, "0x")
}

// Add missing PendingUser model
type PendingUser struct {
	ID                int             `json:"id"`
	ApprovalRequestID int             `json:"approval_request_id"`
	WalletAddress     string          `json:"wallet_address"`
	Username          string          `json:"username"`
	Email             *string         `json:"email"`
	PasswordHash      string          `json:"-"`
	RequestedRole     string          `json:"requested_role"`
	CompanyName       *string         `json:"company_name"`
	CompanyType       *string         `json:"company_type"`
	BusinessLicense   *string         `json:"business_license"`
	ContactInfo       *models.JSONMap `json:"contact_info"`
	Justification     *string         `json:"justification"`
	CreatedAt         time.Time       `json:"created_at"`
}
