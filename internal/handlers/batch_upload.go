package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"aluminium-passport/internal/models"
	"aluminium-passport/internal/services"
	"aluminium-passport/internal/utils"
)

const (
	MaxUploadSize = 50 << 20 // 50MB
)

// BatchUploadHandler handles ZIP file uploads containing multiple passport data
func BatchUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Set max upload size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Parse multipart form
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("zip_file")
	if err != nil {
		http.Error(w, "No file provided or invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !isZipFile(fileHeader.Filename) {
		http.Error(w, "Only ZIP files are allowed", http.StatusBadRequest)
		return
	}

	// Get user info for audit logging
	user, role := extractUserRole(r)

	// Process ZIP file
	processor := utils.NewZipProcessor()
	response, err := processor.ProcessZipFile(fileHeader)
	if err != nil {
		services.LogEvent(user, role, "BATCH_UPLOAD_FAILED", fmt.Sprintf("Error: %v", err))
		http.Error(w, fmt.Sprintf("Failed to process ZIP file: %v", err), http.StatusInternalServerError)
		return
	}

	// Save passports to database (if any were successfully processed)
	if response.Successful > 0 {
		// Note: This would need to be implemented in services package
		// For now, we'll just log the event
		services.LogEvent(user, role, "BATCH_UPLOAD_SUCCESS",
			fmt.Sprintf("Batch: %s, Processed: %d, Successful: %d, Failed: %d",
				response.BatchID, response.TotalProcessed, response.Successful, response.Failed))

		// Upload batch metadata to IPFS
		ipfsHash, err := utils.UploadToIPFS(response)
		if err == nil {
			response.IPFSHash = ipfsHash
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateZipHandler validates a ZIP file without processing it
func ValidateZipHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("zip_file")
	if err != nil {
		http.Error(w, "No file provided or invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !isZipFile(fileHeader.Filename) {
		http.Error(w, "Only ZIP files are allowed", http.StatusBadRequest)
		return
	}

	// Create validation response
	validation := map[string]interface{}{
		"valid":    true,
		"filename": fileHeader.Filename,
		"size":     fileHeader.Size,
		"max_size": MaxUploadSize,
		"message":  "ZIP file is valid and ready for processing",
	}

	// Basic size validation
	if fileHeader.Size > MaxUploadSize {
		validation["valid"] = false
		validation["message"] = fmt.Sprintf("File too large: %d bytes (max: %d)", fileHeader.Size, MaxUploadSize)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

// BatchStatusHandler returns the status of a batch upload
func BatchStatusHandler(w http.ResponseWriter, r *http.Request) {
	batchID := r.URL.Query().Get("batch_id")
	if batchID == "" {
		http.Error(w, "batch_id parameter is required", http.StatusBadRequest)
		return
	}

	// Get user info for audit logging
	user, role := extractUserRole(r)
	services.LogEvent(user, role, "BATCH_STATUS_CHECK", batchID)

	// This would typically fetch from database
	// For now, return a placeholder response
	status := map[string]interface{}{
		"batch_id":   batchID,
		"status":     "completed", // could be: pending, processing, completed, failed
		"message":    "Batch processing completed successfully",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:01:00Z",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// DownloadBatchTemplateHandler provides CSV/JSON templates for batch uploads
func DownloadBatchTemplateHandler(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	switch strings.ToLower(format) {
	case "csv":
		downloadCSVTemplate(w)
	case "json":
		downloadJSONTemplate(w)
	default:
		http.Error(w, "Supported formats: csv, json", http.StatusBadRequest)
	}
}

func downloadCSVTemplate(w http.ResponseWriter) {
	csvContent := `passport_id,batch_id,bauxite_origin,mine_operator,date_of_extraction,refinery_location,refiner_id,smelting_energy_source,carbon_emissions_per_kg,alloy_composition,trace_metals,manufacturer_id,process_type,manufactured_product,product_weight,energy_used,manufacturing_emissions,transport_mode,distance_travelled,logistics_partner_id,shipment_date,recycled_content_percent,recycling_date,recycler_id,recycling_method,times_recycled,certification_agency,esg_score,compliance_standards,date_of_certification,verifier_signature
PASS001,BATCH001,Australia,BHP Billiton,2024-01-01,Queensland Refinery,REF001,Solar,2.5,Al-99.7%,Fe-0.2%,MFG001,Extrusion,Window Frame,10.5,150.0,1.2,Truck,500.0,LOG001,2024-01-15,25.0,2023-12-01,REC001,Mechanical,3,ASI,85.5,ASI Standard,2024-01-10,SIG001`

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=aluminium_passport_template.csv")
	w.Write([]byte(csvContent))
}

func downloadJSONTemplate(w http.ResponseWriter) {
	template := []models.AluminiumPassport{
		{
			PassportID:             "PASS001",
			BatchID:                "BATCH001",
			BauxiteOrigin:          "Australia",
			MineOperator:           "BHP Billiton",
			DateOfExtraction:       "2024-01-01",
			RefineryLocation:       "Queensland Refinery",
			RefinerID:              "REF001",
			SmeltingEnergySource:   "Solar",
			CarbonEmissionsPerKg:   2.5,
			AlloyComposition:       "Al-99.7%",
			TraceMetals:            "Fe-0.2%",
			ManufacturerID:         "MFG001",
			ProcessType:            "Extrusion",
			ManufacturedProduct:    "Window Frame",
			ProductWeight:          10.5,
			EnergyUsed:             150.0,
			ManufacturingEmissions: 1.2,
			TransportMode:          "Truck",
			DistanceTravelled:      500.0,
			LogisticsPartnerID:     "LOG001",
			ShipmentDate:           "2024-01-15",
			RecycledContentPercent: 25.0,
			RecyclingDate:          "2023-12-01",
			RecyclerID:             "REC001",
			RecyclingMethod:        "Mechanical",
			TimesRecycled:          3,
			CertificationAgency:    "ASI",
			ESGScore:               85.5,
			ComplianceStandards:    "ASI Standard",
			DateOfCertification:    "2024-01-10",
			VerifierSignature:      "SIG001",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=aluminium_passport_template.json")
	json.NewEncoder(w).Encode(template)
}

func isZipFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".zip")
}
