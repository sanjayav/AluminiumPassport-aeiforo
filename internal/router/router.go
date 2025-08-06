package router

import (
	"net/http"

	"aluminium-passport/internal/handlers"
	"aluminium-passport/internal/middleware"
	"aluminium-passport/internal/models"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Authentication
	r.HandleFunc("/api/auth/login", handlers.LoginHandler).Methods("POST")

	// API routes with authentication
	api := r.PathPrefix("/api").Subrouter()

	// Single passport operations
	api.HandleFunc("/passports", middleware.RoleGuard(models.RoleIssuer)(
		http.HandlerFunc(handlers.CreatePassportHandler))).Methods("POST")

	api.HandleFunc("/passports/{id}", middleware.RoleGuard(models.RoleViewer, models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.GetPassportByIdHandler))).Methods("GET")

	// Batch operations - ZIP file upload
	api.HandleFunc("/batch/upload", middleware.RoleGuard(models.RoleIssuer)(
		http.HandlerFunc(handlers.BatchUploadHandler))).Methods("POST")

	api.HandleFunc("/batch/validate", middleware.RoleGuard(models.RoleIssuer, models.RoleAuditor)(
		http.HandlerFunc(handlers.ValidateZipHandler))).Methods("POST")

	api.HandleFunc("/batch/status", middleware.RoleGuard(models.RoleViewer, models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.BatchStatusHandler))).Methods("GET")

	// Template downloads
	api.HandleFunc("/batch/template", middleware.RoleGuard(models.RoleIssuer, models.RoleAuditor)(
		http.HandlerFunc(handlers.DownloadBatchTemplateHandler))).Methods("GET")

	// Export operations
	api.HandleFunc("/export/csv", middleware.RoleGuard(models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.ExportCSVHandler))).Methods("GET")

	api.HandleFunc("/export/json", middleware.RoleGuard(models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.ExportJSONHandler))).Methods("GET")

	// Advanced features
	api.HandleFunc("/verify/signature", middleware.RoleGuard(models.RoleViewer, models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.VerifySignatureHandler))).Methods("POST")

	api.HandleFunc("/generate/qr/{id}", middleware.RoleGuard(models.RoleViewer, models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.GenerateQRHandler))).Methods("GET")

	// Zero-knowledge proof endpoints
	api.HandleFunc("/zk/generate", middleware.RoleGuard(models.RoleIssuer)(
		http.HandlerFunc(handlers.GenerateZKProofHandler))).Methods("POST")

	api.HandleFunc("/zk/verify", middleware.RoleGuard(models.RoleViewer, models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.VerifyZKProofHandler))).Methods("POST")

	// Audit and reporting
	api.HandleFunc("/audit/logs", middleware.RoleGuard(models.RoleAuditor, models.RoleIssuer)(
		http.HandlerFunc(handlers.GetAuditLogsHandler))).Methods("GET")

	return r
}
