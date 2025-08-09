package models

import (
	"time"

	"aluminium-passport/internal/db"
)

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
	ApprovalStatusExpired  ApprovalStatus = "expired"
)

// ApprovalType represents the type of approval request
type ApprovalType string

const (
	ApprovalTypeSupplierOnboarding  ApprovalType = "supplier_onboarding"
	ApprovalTypeUserRoleChange      ApprovalType = "user_role_change"
	ApprovalTypeSystemConfiguration ApprovalType = "system_configuration"
)

// ApprovalRequest represents a request that needs approval
type ApprovalRequest struct {
	ID              int            `json:"id" db:"id"`
	RequestType     ApprovalType   `json:"request_type" db:"request_type"`
	RequestedBy     int            `json:"requested_by" db:"requested_by"`
	RequestedByUser *db.User       `json:"requested_by_user,omitempty"`
	ApproverRole    string         `json:"approver_role" db:"approver_role"`
	ApprovedBy      *int           `json:"approved_by" db:"approved_by"`
	ApprovedByUser  *db.User       `json:"approved_by_user,omitempty"`
	Status          ApprovalStatus `json:"status" db:"status"`
	Title           string         `json:"title" db:"title"`
	Description     string         `json:"description" db:"description"`
	RequestData     *db.JSONMap    `json:"request_data" db:"request_data"`
	ApprovalReason  *string        `json:"approval_reason" db:"approval_reason"`
	RejectionReason *string        `json:"rejection_reason" db:"rejection_reason"`
	ExpiresAt       *time.Time     `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	ApprovedAt      *time.Time     `json:"approved_at" db:"approved_at"`
}

// SupplierOnboardingRequest represents the data for supplier onboarding
type SupplierOnboardingRequest struct {
	WalletAddress   string     `json:"wallet_address"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	CompanyName     string     `json:"company_name"`
	CompanyType     string     `json:"company_type"`
	BusinessLicense string     `json:"business_license"`
	ContactInfo     db.JSONMap `json:"contact_info"`
	Justification   string     `json:"justification"`
}

// ApprovalRequestCreateData represents data needed to create an approval request
type ApprovalRequestCreateData struct {
	RequestType   ApprovalType `json:"request_type" binding:"required"`
	ApproverRole  string       `json:"approver_role" binding:"required"`
	Title         string       `json:"title" binding:"required"`
	Description   string       `json:"description" binding:"required"`
	RequestData   *db.JSONMap  `json:"request_data"`
	ExpiresInDays int          `json:"expires_in_days"` // Default 7 days if not specified
}

// ApprovalAction represents an approval or rejection action
type ApprovalAction struct {
	Action ApprovalStatus `json:"action" binding:"required"` // "approved" or "rejected"
	Reason string         `json:"reason"`
}

// IsExpired checks if the approval request has expired
func (ar *ApprovalRequest) IsExpired() bool {
	if ar.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*ar.ExpiresAt)
}

// CanBeApprovedBy checks if the given user role can approve this request
func (ar *ApprovalRequest) CanBeApprovedBy(userRole string) bool {
	return HasHigherOrEqualRole(userRole, ar.ApproverRole)
}

// IsPending checks if the request is still pending
func (ar *ApprovalRequest) IsPending() bool {
	return ar.Status == ApprovalStatusPending && !ar.IsExpired()
}
