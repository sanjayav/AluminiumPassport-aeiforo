package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"aluminium-passport/internal/models"
	"aluminium-passport/internal/utils"
)

// BatchService handles batch operations for passport processing
type BatchService struct {
	db *sql.DB
}

func NewBatchService(db *sql.DB) *BatchService {
	return &BatchService{db: db}
}

// ProcessBatchPassports saves multiple passports to database
func (bs *BatchService) ProcessBatchPassports(passports []models.AluminiumPassport) (*models.BatchUploadResponse, error) {
	if len(passports) == 0 {
		return nil, fmt.Errorf("no passports to process")
	}

	batchID := utils.GenerateBatchID()
	successful := 0
	var errors []string

	// Begin transaction
	tx, err := bs.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO aluminium_passports (
			passport_id, batch_id, bauxite_origin, mine_operator, date_of_extraction,
			refinery_location, refiner_id, smelting_energy_source, carbon_emissions_per_kg,
			alloy_composition, trace_metals, manufacturer_id, process_type, manufactured_product,
			product_weight, energy_used, manufacturing_emissions, transport_mode, distance_travelled,
			logistics_partner_id, shipment_date, recycled_content_percent, recycling_date,
			recycler_id, recycling_method, times_recycled, certification_agency, esg_score,
			compliance_standards, date_of_certification, verifier_signature
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Process each passport
	for i, passport := range passports {
		if passport.BatchID == "" {
			passport.BatchID = batchID
		}

		_, err := stmt.Exec(
			passport.PassportID, passport.BatchID, passport.BauxiteOrigin, passport.MineOperator,
			passport.DateOfExtraction, passport.RefineryLocation, passport.RefinerID,
			passport.SmeltingEnergySource, passport.CarbonEmissionsPerKg, passport.AlloyComposition,
			passport.TraceMetals, passport.ManufacturerID, passport.ProcessType, passport.ManufacturedProduct,
			passport.ProductWeight, passport.EnergyUsed, passport.ManufacturingEmissions, passport.TransportMode,
			passport.DistanceTravelled, passport.LogisticsPartnerID, passport.ShipmentDate,
			passport.RecycledContentPercent, passport.RecyclingDate, passport.RecyclerID, passport.RecyclingMethod,
			passport.TimesRecycled, passport.CertificationAgency, passport.ESGScore, passport.ComplianceStandards,
			passport.DateOfCertification, passport.VerifierSignature,
		)

		if err != nil {
			errors = append(errors, fmt.Sprintf("Passport %d (%s): %v", i+1, passport.PassportID, err))
		} else {
			successful++
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Upload to IPFS if successful
	var ipfsHash string
	if successful > 0 {
		ipfsData := map[string]interface{}{
			"batch_id":     batchID,
			"passports":    passports[:successful],
			"processed_at": time.Now().UTC(),
		}
		hash, err := utils.UploadToIPFS(ipfsData)
		if err == nil {
			ipfsHash = hash
		}
	}

	response := &models.BatchUploadResponse{
		BatchID:        batchID,
		TotalProcessed: len(passports),
		Successful:     successful,
		Failed:         len(errors),
		Errors:         errors,
		IPFSHash:       ipfsHash,
	}

	return response, nil
}

// GetBatchStatus returns the status of a batch
func (bs *BatchService) GetBatchStatus(batchID string) (map[string]interface{}, error) {
	var count int
	err := bs.db.QueryRow("SELECT COUNT(*) FROM aluminium_passports WHERE batch_id = $1", batchID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch status: %w", err)
	}

	status := map[string]interface{}{
		"batch_id":    batchID,
		"status":      "completed",
		"total_count": count,
		"created_at":  time.Now().UTC(),
	}

	return status, nil
}

// ExportPassportsCSV exports passports as CSV
func ExportPassportsCSV(batchID string) ([]byte, error) {
	// This would query the database and format as CSV
	// For now, return a placeholder
	csvContent := "passport_id,batch_id,bauxite_origin,mine_operator\n"
	csvContent += "PASS001,BATCH001,Australia,BHP Billiton\n"

	return []byte(csvContent), nil
}

// ExportPassportsJSON exports passports as JSON
func ExportPassportsJSON(batchID string) ([]byte, error) {
	// This would query the database and format as JSON
	// For now, return a placeholder
	passports := []models.AluminiumPassport{
		{
			PassportID:    "PASS001",
			BatchID:       batchID,
			BauxiteOrigin: "Australia",
			MineOperator:  "BHP Billiton",
		},
	}

	return json.Marshal(passports)
}

// GetAuditLogs returns audit logs with filtering
func GetAuditLogs(userFilter, actionFilter, limit string) ([]map[string]interface{}, error) {
	// This would query the audit log table
	// For now, return placeholder data
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().UTC(),
			"user":      "admin",
			"role":      "issuer",
			"action":    "BATCH_UPLOAD",
			"target":    "BATCH001",
		},
	}

	return logs, nil
}
