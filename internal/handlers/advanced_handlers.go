package handlers

import (
	"encoding/json"
	"net/http"

	"aluminium-passport/internal/services"
	"aluminium-passport/internal/utils"

	"github.com/gorilla/mux"
)

// VerifySignatureHandler verifies digital signatures
func VerifySignatureHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Data      string `json:"data"`
		Signature string `json:"signature"`
		PublicKey string `json:"public_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	isValid := services.VerifySignature(request.Data, request.Signature, request.PublicKey)

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "SIGNATURE_VERIFY", request.Signature)

	response := map[string]interface{}{
		"valid":   isValid,
		"message": "Signature verification completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateQRHandler generates QR codes for passport data
func GenerateQRHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	passport, err := services.GetPassportById(id)
	if err != nil {
		http.Error(w, "Passport not found", http.StatusNotFound)
		return
	}

	qrCode, err := utils.GenerateQRCode(passport)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "QR_GENERATE", id)

	w.Header().Set("Content-Type", "image/png")
	w.Write(qrCode)
}

// GenerateZKProofHandler generates zero-knowledge proofs
func GenerateZKProofHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PassportID string                 `json:"passport_id"`
		Claims     map[string]interface{} `json:"claims"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	proof, err := services.GenerateZKProof(request.PassportID, request.Claims)
	if err != nil {
		http.Error(w, "Failed to generate ZK proof", http.StatusInternalServerError)
		return
	}

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "ZK_PROOF_GENERATE", request.PassportID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"proof":   proof,
		"message": "ZK proof generated successfully",
	})
}

// VerifyZKProofHandler verifies zero-knowledge proofs
func VerifyZKProofHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Proof  string                 `json:"proof"`
		Claims map[string]interface{} `json:"claims"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	isValid, err := services.VerifyZKProof(request.Proof, request.Claims)
	if err != nil {
		http.Error(w, "Failed to verify ZK proof", http.StatusInternalServerError)
		return
	}

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "ZK_PROOF_VERIFY", request.Proof)

	response := map[string]interface{}{
		"valid":   isValid,
		"message": "ZK proof verification completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAuditLogsHandler returns audit logs
func GetAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Query parameters for filtering
	userFilter := r.URL.Query().Get("user")
	actionFilter := r.URL.Query().Get("action")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}

	logs, err := services.GetAuditLogs(userFilter, actionFilter, limit)
	if err != nil {
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "AUDIT_LOG_ACCESS", "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// ExportCSVHandler exports passport data as CSV
func ExportCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Query parameters for filtering
	batchID := r.URL.Query().Get("batch_id")

	csvData, err := services.ExportPassportsCSV(batchID)
	if err != nil {
		http.Error(w, "Failed to export CSV", http.StatusInternalServerError)
		return
	}

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "EXPORT_CSV", batchID)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=passports_export.csv")
	w.Write(csvData)
}

// ExportJSONHandler exports passport data as JSON
func ExportJSONHandler(w http.ResponseWriter, r *http.Request) {
	// Query parameters for filtering
	batchID := r.URL.Query().Get("batch_id")

	jsonData, err := services.ExportPassportsJSON(batchID)
	if err != nil {
		http.Error(w, "Failed to export JSON", http.StatusInternalServerError)
		return
	}

	user, role := extractUserRole(r)
	services.LogEvent(user, role, "EXPORT_JSON", batchID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=passports_export.json")
	w.Write(jsonData)
}
