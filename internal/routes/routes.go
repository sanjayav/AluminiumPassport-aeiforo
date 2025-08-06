package routes

import (
	"net/http"

	"aluminium-passport/internal/controller"
	"aluminium-passport/internal/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Initialize controllers
	authController := controller.NewAuthController()
	passportController := controller.NewPassportController()
	esgController := controller.NewESGController()
	approvalController := controller.NewApprovalController()

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "aluminium-passport-api"}`))
	}).Methods("GET")

	// Public endpoints (no authentication required)
	public := r.PathPrefix("/api/public").Subrouter()
	public.HandleFunc("/verify/{id}", passportController.GetPassportDetails).Methods("GET")
	public.HandleFunc("/qr/{id}", passportController.GetQRCode).Methods("GET")

	// Authentication endpoints
	auth := r.PathPrefix("/api/auth").Subrouter()
	auth.HandleFunc("/login", authController.Login).Methods("POST")
	auth.HandleFunc("/register", authController.Register).Methods("POST")
	auth.HandleFunc("/refresh", authController.RefreshToken).Methods("POST")
	auth.HandleFunc("/logout", middleware.AuthMiddleware(authController.Logout)).Methods("POST")
	auth.HandleFunc("/profile", middleware.AuthMiddleware(authController.GetProfile)).Methods("GET")

	// Protected API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// Passport management routes
	passports := api.PathPrefix("/passports").Subrouter()

	// Create passport (miners, manufacturers, admins)
	passports.HandleFunc("", middleware.RoleMiddleware("miner", "manufacturer", "admin")(
		passportController.RegisterPassport)).Methods("POST")

	// Get passport details (all authenticated users)
	passports.HandleFunc("/{id}", passportController.GetPassportDetails).Methods("GET")

	// Update recycled content (recyclers, admins)
	passports.HandleFunc("/{id}/recycle", middleware.RoleMiddleware("recycler", "admin")(
		passportController.UpdateRecycledContent)).Methods("PUT")

	// Get QR code (all authenticated users)
	passports.HandleFunc("/{id}/qr", passportController.GetQRCode).Methods("GET")

	// List passports with pagination (all authenticated users)
	passports.HandleFunc("", passportController.ListPassports).Methods("GET")

	// ESG management routes
	esg := api.PathPrefix("/esg").Subrouter()

	// Create ESG assessment (certifiers, admins)
	esg.HandleFunc("/assess", middleware.RoleMiddleware("certifier", "admin")(
		esgController.CreateESGAssessment)).Methods("POST")

	// Get ESG metrics (all authenticated users)
	esg.HandleFunc("/{id}", esgController.GetESGMetrics).Methods("GET")

	// Generate AI-based ESG score (certifiers, admins)
	esg.HandleFunc("/generate", middleware.RoleMiddleware("certifier", "admin")(
		esgController.GenerateESGScore)).Methods("POST")

	// Get ESG ranking (all authenticated users)
	esg.HandleFunc("/ranking", esgController.GetESGRanking).Methods("GET")

	// Batch operations routes
	batch := api.PathPrefix("/batch").Subrouter()

	// Upload ZIP file (miners, manufacturers, admins)
	batch.HandleFunc("/upload", middleware.RoleMiddleware("miner", "manufacturer", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for batch upload
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "Batch upload functionality to be implemented"}`))
		})).Methods("POST")

	// Validate ZIP file (miners, manufacturers, certifiers, admins)
	batch.HandleFunc("/validate", middleware.RoleMiddleware("miner", "manufacturer", "certifier", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for batch validation
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "Batch validation functionality to be implemented"}`))
		})).Methods("POST")

	// Get batch status (all authenticated users)
	batch.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for batch status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "Batch status functionality to be implemented"}`))
	}).Methods("GET")

	// Export routes
	export := api.PathPrefix("/export").Subrouter()

	// Export CSV (auditors, certifiers, admins)
	export.HandleFunc("/csv", middleware.RoleMiddleware("auditor", "certifier", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for CSV export
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", "attachment; filename=passports.csv")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte("passport_id,manufacturer,origin,esg_score\n"))
		})).Methods("GET")

	// Export JSON (auditors, certifiers, admins)
	export.HandleFunc("/json", middleware.RoleMiddleware("auditor", "certifier", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for JSON export
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "JSON export functionality to be implemented"}`))
		})).Methods("GET")

	// Verification routes
	verify := api.PathPrefix("/verify").Subrouter()

	// Verify signature (all authenticated users)
	verify.HandleFunc("/signature", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for signature verification
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "Signature verification functionality to be implemented"}`))
	}).Methods("POST")

	// Zero-knowledge proof routes
	zk := api.PathPrefix("/zk").Subrouter()

	// Generate ZK proof (certifiers, admins)
	zk.HandleFunc("/generate", middleware.RoleMiddleware("certifier", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for ZK proof generation
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "ZK proof generation functionality to be implemented"}`))
		})).Methods("POST")

	// Verify ZK proof (all authenticated users)
	zk.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for ZK proof verification
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "ZK proof verification functionality to be implemented"}`))
	}).Methods("POST")

	// Audit routes
	audit := api.PathPrefix("/audit").Subrouter()

	// Get audit logs (auditors, admins)
	audit.HandleFunc("/logs", middleware.RoleMiddleware("auditor", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for audit logs
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "Audit logs functionality to be implemented"}`))
		})).Methods("GET")

	// Blockchain integration routes
	blockchain := api.PathPrefix("/blockchain").Subrouter()

	// Register passport on blockchain (miners, manufacturers, admins)
	blockchain.HandleFunc("/register/{id}", middleware.RoleMiddleware("miner", "manufacturer", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for blockchain registration
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "Blockchain registration functionality to be implemented"}`))
		})).Methods("POST")

	// Get blockchain transaction (all authenticated users)
	blockchain.HandleFunc("/tx/{hash}", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for blockchain transaction info
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "Blockchain transaction info functionality to be implemented"}`))
	}).Methods("GET")

	// IPFS routes
	ipfs := api.PathPrefix("/ipfs").Subrouter()

	// Upload to IPFS (miners, manufacturers, certifiers, admins)
	ipfs.HandleFunc("/upload/{id}", middleware.RoleMiddleware("miner", "manufacturer", "certifier", "admin")(
		func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for IPFS upload
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"message": "IPFS upload functionality to be implemented"}`))
		})).Methods("POST")

	// Get from IPFS (all authenticated users)
	ipfs.HandleFunc("/{hash}", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for IPFS retrieval
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "IPFS retrieval functionality to be implemented"}`))
	}).Methods("GET")

	// Approval workflow routes
	approvals := api.PathPrefix("/approvals").Subrouter()

	// Create approval request (admins and above)
	approvals.HandleFunc("", middleware.RoleMiddleware("admin", "super_admin")(
		approvalController.CreateApprovalRequest)).Methods("POST")

	// Request supplier onboarding (admins only - requires super admin approval)
	approvals.HandleFunc("/supplier-onboarding", middleware.RoleMiddleware("admin")(
		approvalController.RequestSupplierOnboarding)).Methods("POST")

	// Get approval requests (all authenticated users can see their own, approvers can see pending)
	approvals.HandleFunc("", approvalController.GetApprovalRequests).Methods("GET")

	// Get specific approval request
	approvals.HandleFunc("/{id}", approvalController.GetApprovalRequest).Methods("GET")

	// Approve request (super admins can approve supplier onboarding)
	approvals.HandleFunc("/{id}/approve", middleware.RoleMiddleware("super_admin")(
		approvalController.ApproveRequest)).Methods("POST")

	// Reject request (super admins can reject supplier onboarding)
	approvals.HandleFunc("/{id}/reject", middleware.RoleMiddleware("super_admin")(
		approvalController.RejectRequest)).Methods("POST")

	// Admin routes (requires admin or super_admin)
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RoleMiddleware("admin", "super_admin"))

	// User management
	admin.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for user management
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "User management functionality to be implemented"}`))
	}).Methods("GET")

	// System statistics
	admin.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for system stats
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "System statistics functionality to be implemented"}`))
	}).Methods("GET")

	// Super Admin only routes
	superAdmin := api.PathPrefix("/super-admin").Subrouter()
	superAdmin.Use(middleware.RoleMiddleware("super_admin"))

	// Manage admin users
	superAdmin.HandleFunc("/admins", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "Admin user management functionality to be implemented"}`))
	}).Methods("GET")

	// System configuration
	superAdmin.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"message": "System configuration functionality to be implemented"}`))
	}).Methods("GET", "POST")

	// Apply CORS middleware to all routes
	r.Use(middleware.CORSMiddleware)

	// Apply logging middleware
	r.Use(middleware.LoggingMiddleware)

	// Apply rate limiting middleware
	r.Use(middleware.RateLimitMiddleware)

	return r
}
