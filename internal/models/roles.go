package models

const (
	// Hierarchical roles - highest to lowest privilege
	RoleSuperAdmin   = "super_admin"  // Highest level - can approve admin actions
	RoleAdmin        = "admin"        // System admin - needs super admin approval for supplier onboarding
	RoleCertifier    = "certifier"    // Can create ESG assessments and certifications
	RoleAuditor      = "auditor"      // Can view audit logs and export data
	RoleMiner        = "miner"        // Can create passports from mining operations
	RoleManufacturer = "manufacturer" // Can create passports from manufacturing
	RoleRecycler     = "recycler"     // Can update recycling information
	RoleViewer       = "viewer"       // Basic read-only access

	// Legacy roles (maintained for backward compatibility)
	RoleIssuer = "issuer" // Deprecated: use RoleMiner or RoleManufacturer
)

// Role hierarchy levels (higher number = higher privilege)
var RoleHierarchy = map[string]int{
	RoleSuperAdmin:   100,
	RoleAdmin:        90,
	RoleCertifier:    80,
	RoleAuditor:      70,
	RoleMiner:        60,
	RoleManufacturer: 60,
	RoleRecycler:     50,
	RoleViewer:       10,
	RoleIssuer:       60, // Legacy compatibility
}

// GetRoleLevel returns the hierarchy level for a role
func GetRoleLevel(role string) int {
	if level, exists := RoleHierarchy[role]; exists {
		return level
	}
	return 0 // Unknown role gets lowest level
}

// HasHigherOrEqualRole checks if userRole has higher or equal privilege than requiredRole
func HasHigherOrEqualRole(userRole, requiredRole string) bool {
	return GetRoleLevel(userRole) >= GetRoleLevel(requiredRole)
}

// CanApprove checks if approverRole can approve actions by targetRole
func CanApprove(approverRole, targetRole string) bool {
	return GetRoleLevel(approverRole) > GetRoleLevel(targetRole)
}

// GetValidRoles returns all valid roles
func GetValidRoles() []string {
	return []string{
		RoleSuperAdmin,
		RoleAdmin,
		RoleCertifier,
		RoleAuditor,
		RoleMiner,
		RoleManufacturer,
		RoleRecycler,
		RoleViewer,
	}
}

// GetSupplierRoles returns roles that are considered suppliers
func GetSupplierRoles() []string {
	return []string{
		RoleMiner,
		RoleManufacturer,
		RoleRecycler,
		RoleCertifier,
	}
}

// RequiresSuperAdminApproval checks if onboarding this role requires super admin approval
func RequiresSuperAdminApproval(role string) bool {
	supplierRoles := GetSupplierRoles()
	for _, supplierRole := range supplierRoles {
		if role == supplierRole {
			return true
		}
	}
	return false
}
