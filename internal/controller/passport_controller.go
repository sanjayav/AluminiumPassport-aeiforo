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
	"aluminium-passport/internal/ipfs"
	"aluminium-passport/internal/qr"

	"github.com/gorilla/mux"
)

type PassportController struct{}

func NewPassportController() *PassportController {
	return &PassportController{}
}

type CreatePassportRequest struct {
	PassportID             string      `json:"passport_id" binding:"required"`
	BatchID                *string     `json:"batch_id"`
	Manufacturer           string      `json:"manufacturer" binding:"required"`
	Origin                 string      `json:"origin" binding:"required"`
	BauxiteSource          *string     `json:"bauxite_source"`
	AlloyComposition       *string     `json:"alloy_composition"`
	MineOperator           *string     `json:"mine_operator"`
	DateOfExtraction       *string     `json:"date_of_extraction"`
	ExtractionMethod       *string     `json:"extraction_method"`
	MineLocation           *string     `json:"mine_location"`
	RefineryLocation       *string     `json:"refinery_location"`
	RefinerID              *string     `json:"refiner_id"`
	RefiningDate           *string     `json:"refining_date"`
	RefiningMethod         *string     `json:"refining_method"`
	SmeltingLocation       *string     `json:"smelting_location"`
	SmeltingEnergySource   *string     `json:"smelting_energy_source"`
	ProcessType            *string     `json:"process_type"`
	ManufacturedProduct    *string     `json:"manufactured_product"`
	ManufacturingDate      *string     `json:"manufacturing_date"`
	ProductWeight          *float64    `json:"product_weight"`
	EnergyUsed             *float64    `json:"energy_used"`
	WaterUsed              *float64    `json:"water_used"`
	WasteGenerated         *float64    `json:"waste_generated"`
	CarbonEmissionsPerKg   *float64    `json:"carbon_emissions_per_kg"`
	CO2Footprint           *float64    `json:"co2_footprint"`
	ManufacturingEmissions *float64    `json:"manufacturing_emissions"`
	TransportMode          *string     `json:"transport_mode"`
	DistanceTravelled      *float64    `json:"distance_travelled"`
	LogisticsPartnerID     *string     `json:"logistics_partner_id"`
	ShipmentDate           *string     `json:"shipment_date"`
	RecycledContentPercent *float64    `json:"recycled_content_percent"`
	RecyclingDate          *string     `json:"recycling_date"`
	RecyclerID             *string     `json:"recycler_id"`
	RecyclingMethod        *string     `json:"recycling_method"`
	TimesRecycled          *int        `json:"times_recycled"`
	CertificationAgency    *string     `json:"certification_agency"`
	Certifier              *string     `json:"certifier"`
	ComplianceStandards    *string     `json:"compliance_standards"`
	DateOfCertification    *string     `json:"date_of_certification"`
	CertificationExpiry    *string     `json:"certification_expiry"`
	VerifierSignature      *string     `json:"verifier_signature"`
	Metadata               *db.JSONMap `json:"metadata"`
}

type UpdatePassportRequest struct {
	RecycledContentPercent *float64    `json:"recycled_content_percent"`
	RecyclingMethod        *string     `json:"recycling_method"`
	TimesRecycled          *int        `json:"times_recycled"`
	ESGScore               *float64    `json:"esg_score"`
	EnvironmentalScore     *float64    `json:"environmental_score"`
	SocialScore            *float64    `json:"social_score"`
	GovernanceScore        *float64    `json:"governance_score"`
	Metadata               *db.JSONMap `json:"metadata"`
}

type PassportResponse struct {
	*db.AluminiumPassport
	QRCodeURL string `json:"qr_code_url,omitempty"`
	IPFSUrl   string `json:"ipfs_url,omitempty"`
}

// RegisterPassport creates a new passport
func (pc *PassportController) RegisterPassport(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	claims, err := pc.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has miner or manufacturer role
	if !pc.hasRole(claims.Role, []string{"miner", "manufacturer", "admin"}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req CreatePassportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if passport ID already exists
	if exists, err := pc.passportExists(req.PassportID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if exists {
		http.Error(w, "Passport ID already exists", http.StatusConflict)
		return
	}

	// Create passport object
	passport := &db.AluminiumPassport{
		PassportID:             req.PassportID,
		BatchID:                req.BatchID,
		Manufacturer:           req.Manufacturer,
		Origin:                 req.Origin,
		BauxiteSource:          req.BauxiteSource,
		AlloyComposition:       req.AlloyComposition,
		MineOperator:           req.MineOperator,
		ExtractionMethod:       req.ExtractionMethod,
		MineLocation:           req.MineLocation,
		RefineryLocation:       req.RefineryLocation,
		RefinerID:              req.RefinerID,
		RefiningMethod:         req.RefiningMethod,
		SmeltingLocation:       req.SmeltingLocation,
		SmeltingEnergySource:   req.SmeltingEnergySource,
		ProcessType:            req.ProcessType,
		ManufacturedProduct:    req.ManufacturedProduct,
		ProductWeight:          req.ProductWeight,
		EnergyUsed:             req.EnergyUsed,
		WaterUsed:              req.WaterUsed,
		WasteGenerated:         req.WasteGenerated,
		CarbonEmissionsPerKg:   req.CarbonEmissionsPerKg,
		CO2Footprint:           req.CO2Footprint,
		ManufacturingEmissions: req.ManufacturingEmissions,
		TransportMode:          req.TransportMode,
		DistanceTravelled:      req.DistanceTravelled,
		LogisticsPartnerID:     req.LogisticsPartnerID,
		RecycledContentPercent: req.RecycledContentPercent,
		RecyclerID:             req.RecyclerID,
		RecyclingMethod:        req.RecyclingMethod,
		TimesRecycled:          getIntValue(req.TimesRecycled, 0),
		CertificationAgency:    req.CertificationAgency,
		Certifier:              req.Certifier,
		ComplianceStandards:    req.ComplianceStandards,
		VerifierSignature:      req.VerifierSignature,
		Metadata:               req.Metadata,
		Status:                 "active",
		IsVerified:             false,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		CreatedBy:              &claims.UserID,
		UpdatedBy:              &claims.UserID,
	}

	// Parse date fields
	if req.DateOfExtraction != nil {
		if date, err := time.Parse("2006-01-02", *req.DateOfExtraction); err == nil {
			passport.DateOfExtraction = &date
		}
	}
	if req.RefiningDate != nil {
		if date, err := time.Parse("2006-01-02", *req.RefiningDate); err == nil {
			passport.RefiningDate = &date
		}
	}
	if req.ManufacturingDate != nil {
		if date, err := time.Parse("2006-01-02", *req.ManufacturingDate); err == nil {
			passport.ManufacturingDate = &date
		}
	}
	if req.ShipmentDate != nil {
		if date, err := time.Parse("2006-01-02", *req.ShipmentDate); err == nil {
			passport.ShipmentDate = &date
		}
	}
	if req.RecyclingDate != nil {
		if date, err := time.Parse("2006-01-02", *req.RecyclingDate); err == nil {
			passport.RecyclingDate = &date
		}
	}
	if req.DateOfCertification != nil {
		if date, err := time.Parse("2006-01-02", *req.DateOfCertification); err == nil {
			passport.DateOfCertification = &date
		}
	}
	if req.CertificationExpiry != nil {
		if date, err := time.Parse("2006-01-02", *req.CertificationExpiry); err == nil {
			passport.CertificationExpiry = &date
		}
	}

	// Save to database
	passportID, err := pc.createPassport(passport)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create passport: %v", err), http.StatusInternalServerError)
		return
	}

	passport.ID = passportID

	// Upload to IPFS
	if ipfsHash, err := ipfs.UploadPassportData(passport); err == nil {
		passport.IPFSHash = &ipfsHash
		pc.updatePassportIPFS(passport.PassportID, ipfsHash)
	}

	// Generate QR code
	if qrData, err := qr.GeneratePassportQR(passport); err == nil {
		passport.QRCodeData = &qrData
		pc.updatePassportQR(passport.PassportID, qrData)
	}

	// Log audit event
	pc.logAuditEvent(claims.UserID, claims.Role, "CREATE", "passport", passport.PassportID, nil, passport, r)

	// Prepare response
	response := &PassportResponse{
		AluminiumPassport: passport,
	}

	if passport.QRCodeData != nil {
		response.QRCodeURL = fmt.Sprintf("/api/passports/%s/qr", passport.PassportID)
	}

	if passport.IPFSHash != nil {
		response.IPFSUrl = fmt.Sprintf("https://gateway.pinata.cloud/ipfs/%s", strings.TrimPrefix(*passport.IPFSHash, "ipfs://"))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPassportDetails retrieves passport by ID
func (pc *PassportController) GetPassportDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	passportID := vars["id"]

	// Extract user info from token
	claims, err := pc.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get passport from database
	passport, err := pc.getPassportByID(passportID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Passport not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if passport is active
	if passport.Status != "active" {
		http.Error(w, "Passport is not active", http.StatusForbidden)
		return
	}

	// Log audit event
	pc.logAuditEvent(claims.UserID, claims.Role, "VIEW", "passport", passportID, nil, nil, r)

	// Prepare response
	response := &PassportResponse{
		AluminiumPassport: passport,
	}

	if passport.QRCodeData != nil {
		response.QRCodeURL = fmt.Sprintf("/api/passports/%s/qr", passport.PassportID)
	}

	if passport.IPFSHash != nil {
		response.IPFSUrl = fmt.Sprintf("https://gateway.pinata.cloud/ipfs/%s", strings.TrimPrefix(*passport.IPFSHash, "ipfs://"))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateRecycledContent updates recycling information
func (pc *PassportController) UpdateRecycledContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	passportID := vars["id"]

	// Extract user info from token
	claims, err := pc.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has recycler role
	if !pc.hasRole(claims.Role, []string{"recycler", "admin"}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req UpdatePassportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing passport
	passport, err := pc.getPassportByID(passportID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Passport not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store old values for audit
	oldValues := &db.JSONMap{
		"recycled_content_percent": passport.RecycledContentPercent,
		"recycling_method":         passport.RecyclingMethod,
		"times_recycled":           passport.TimesRecycled,
	}

	// Update fields
	updateFields := make(map[string]interface{})
	if req.RecycledContentPercent != nil {
		passport.RecycledContentPercent = req.RecycledContentPercent
		updateFields["recycled_content_percent"] = *req.RecycledContentPercent
	}
	if req.RecyclingMethod != nil {
		passport.RecyclingMethod = req.RecyclingMethod
		updateFields["recycling_method"] = *req.RecyclingMethod
	}
	if req.TimesRecycled != nil {
		passport.TimesRecycled = *req.TimesRecycled
		updateFields["times_recycled"] = *req.TimesRecycled
	}

	// Add recycling timestamp
	now := time.Now()
	passport.LastRecyclingDate = &now
	updateFields["last_recycling_date"] = now

	// Update in database
	if err := pc.updatePassportFields(passportID, updateFields, claims.UserID); err != nil {
		http.Error(w, "Failed to update passport", http.StatusInternalServerError)
		return
	}

	// Log audit event
	newValues := &db.JSONMap(updateFields)
	pc.logAuditEvent(claims.UserID, claims.Role, "UPDATE", "passport", passportID, oldValues, newValues, r)

	// Add supply chain step
	pc.addSupplyChainStep(passportID, "Recycling", fmt.Sprintf("Recycled content updated to %.2f%%", getFloatValue(req.RecycledContentPercent, 0)), claims.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Recycled content updated successfully",
	})
}

// GetQRCode returns the QR code for a passport
func (pc *PassportController) GetQRCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	passportID := vars["id"]

	// Get passport from database
	passport, err := pc.getPassportByID(passportID)
	if err != nil {
		http.Error(w, "Passport not found", http.StatusNotFound)
		return
	}

	// Generate or retrieve QR code
	var qrCode []byte
	if passport.QRCodeData != nil {
		// QR code exists, decode it
		qrCode = []byte(*passport.QRCodeData) // This would need proper base64 decoding
	} else {
		// Generate new QR code
		qrCode, err = qr.GenerateQRCodeImage(passport)
		if err != nil {
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(qrCode)
}

// ListPassports returns a paginated list of passports
func (pc *PassportController) ListPassports(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	claims, err := pc.extractUserClaims(r)
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

	manufacturer := r.URL.Query().Get("manufacturer")
	status := r.URL.Query().Get("status")
	batchID := r.URL.Query().Get("batch_id")

	// Get passports from database
	passports, total, err := pc.listPassports(page, limit, manufacturer, status, batchID)
	if err != nil {
		http.Error(w, "Failed to retrieve passports", http.StatusInternalServerError)
		return
	}

	// Log audit event
	pc.logAuditEvent(claims.UserID, claims.Role, "LIST", "passport", "", nil, nil, r)

	response := map[string]interface{}{
		"passports":   passports,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Database helper methods
func (pc *PassportController) passportExists(passportID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM aluminium_passports WHERE passport_id = $1)`
	var exists bool
	err := db.DB.QueryRow(query, passportID).Scan(&exists)
	return exists, err
}

func (pc *PassportController) createPassport(passport *db.AluminiumPassport) (int, error) {
	query := `
		INSERT INTO aluminium_passports (
			passport_id, batch_id, manufacturer, origin, bauxite_source, alloy_composition,
			mine_operator, date_of_extraction, extraction_method, mine_location,
			refinery_location, refiner_id, refining_date, refining_method,
			smelting_location, smelting_energy_source, process_type, manufactured_product, manufacturing_date,
			product_weight, energy_used, water_used, waste_generated,
			carbon_emissions_per_kg, co2_footprint, manufacturing_emissions,
			transport_mode, distance_travelled, logistics_partner_id, shipment_date,
			recycled_content_percent, recycling_date, recycler_id, recycling_method, times_recycled,
			certification_agency, certifier, compliance_standards, date_of_certification, certification_expiry, verifier_signature,
			metadata, status, is_verified, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35,
			$36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48
		) RETURNING id`

	var passportID int
	err := db.DB.QueryRow(
		query,
		passport.PassportID, passport.BatchID, passport.Manufacturer, passport.Origin, passport.BauxiteSource, passport.AlloyComposition,
		passport.MineOperator, passport.DateOfExtraction, passport.ExtractionMethod, passport.MineLocation,
		passport.RefineryLocation, passport.RefinerID, passport.RefiningDate, passport.RefiningMethod,
		passport.SmeltingLocation, passport.SmeltingEnergySource, passport.ProcessType, passport.ManufacturedProduct, passport.ManufacturingDate,
		passport.ProductWeight, passport.EnergyUsed, passport.WaterUsed, passport.WasteGenerated,
		passport.CarbonEmissionsPerKg, passport.CO2Footprint, passport.ManufacturingEmissions,
		passport.TransportMode, passport.DistanceTravelled, passport.LogisticsPartnerID, passport.ShipmentDate,
		passport.RecycledContentPercent, passport.RecyclingDate, passport.RecyclerID, passport.RecyclingMethod, passport.TimesRecycled,
		passport.CertificationAgency, passport.Certifier, passport.ComplianceStandards, passport.DateOfCertification, passport.CertificationExpiry, passport.VerifierSignature,
		passport.Metadata, passport.Status, passport.IsVerified, passport.CreatedAt, passport.UpdatedAt, passport.CreatedBy, passport.UpdatedBy,
	).Scan(&passportID)

	return passportID, err
}

func (pc *PassportController) getPassportByID(passportID string) (*db.AluminiumPassport, error) {
	query := `
		SELECT id, passport_id, batch_id, manufacturer, origin, bauxite_source, alloy_composition,
		       mine_operator, date_of_extraction, extraction_method, mine_location,
		       refinery_location, refiner_id, refining_date, refining_method,
		       smelting_location, smelting_energy_source, process_type, manufactured_product, manufacturing_date,
		       product_weight, energy_used, water_used, waste_generated,
		       carbon_emissions_per_kg, co2_footprint, manufacturing_emissions,
		       transport_mode, distance_travelled, logistics_partner_id, shipment_date,
		       recycled_content_percent, recycling_date, recycler_id, recycling_method, times_recycled, last_recycling_date,
		       certification_agency, certifier, compliance_standards, date_of_certification, certification_expiry, verifier_signature,
		       esg_score, environmental_score, social_score, governance_score, esg_last_updated,
		       ipfs_hash, qr_code_data, digital_signature,
		       blockchain_tx_hash, contract_address, block_number,
		       status, is_verified, verification_date,
		       metadata, supply_chain_steps, certifications,
		       created_at, updated_at, created_by, updated_by
		FROM aluminium_passports 
		WHERE passport_id = $1`

	passport := &db.AluminiumPassport{}
	err := db.DB.QueryRow(query, passportID).Scan(
		&passport.ID, &passport.PassportID, &passport.BatchID, &passport.Manufacturer, &passport.Origin, &passport.BauxiteSource, &passport.AlloyComposition,
		&passport.MineOperator, &passport.DateOfExtraction, &passport.ExtractionMethod, &passport.MineLocation,
		&passport.RefineryLocation, &passport.RefinerID, &passport.RefiningDate, &passport.RefiningMethod,
		&passport.SmeltingLocation, &passport.SmeltingEnergySource, &passport.ProcessType, &passport.ManufacturedProduct, &passport.ManufacturingDate,
		&passport.ProductWeight, &passport.EnergyUsed, &passport.WaterUsed, &passport.WasteGenerated,
		&passport.CarbonEmissionsPerKg, &passport.CO2Footprint, &passport.ManufacturingEmissions,
		&passport.TransportMode, &passport.DistanceTravelled, &passport.LogisticsPartnerID, &passport.ShipmentDate,
		&passport.RecycledContentPercent, &passport.RecyclingDate, &passport.RecyclerID, &passport.RecyclingMethod, &passport.TimesRecycled, &passport.LastRecyclingDate,
		&passport.CertificationAgency, &passport.Certifier, &passport.ComplianceStandards, &passport.DateOfCertification, &passport.CertificationExpiry, &passport.VerifierSignature,
		&passport.ESGScore, &passport.EnvironmentalScore, &passport.SocialScore, &passport.GovernanceScore, &passport.ESGLastUpdated,
		&passport.IPFSHash, &passport.QRCodeData, &passport.DigitalSignature,
		&passport.BlockchainTxHash, &passport.ContractAddress, &passport.BlockNumber,
		&passport.Status, &passport.IsVerified, &passport.VerificationDate,
		&passport.Metadata, &passport.SupplyChainSteps, &passport.Certifications,
		&passport.CreatedAt, &passport.UpdatedAt, &passport.CreatedBy, &passport.UpdatedBy,
	)

	return passport, err
}

func (pc *PassportController) updatePassportFields(passportID string, fields map[string]interface{}, userID int) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range fields {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	setParts = append(setParts, fmt.Sprintf("updated_by = $%d", argIndex))
	args = append(args, userID)
	argIndex++

	args = append(args, passportID)

	query := fmt.Sprintf("UPDATE aluminium_passports SET %s WHERE passport_id = $%d", strings.Join(setParts, ", "), argIndex)

	_, err := db.DB.Exec(query, args...)
	return err
}

func (pc *PassportController) updatePassportIPFS(passportID, ipfsHash string) error {
	query := `UPDATE aluminium_passports SET ipfs_hash = $1, updated_at = $2 WHERE passport_id = $3`
	_, err := db.DB.Exec(query, ipfsHash, time.Now(), passportID)
	return err
}

func (pc *PassportController) updatePassportQR(passportID, qrData string) error {
	query := `UPDATE aluminium_passports SET qr_code_data = $1, updated_at = $2 WHERE passport_id = $3`
	_, err := db.DB.Exec(query, qrData, time.Now(), passportID)
	return err
}

func (pc *PassportController) listPassports(page, limit int, manufacturer, status, batchID string) ([]*db.AluminiumPassport, int, error) {
	// Build WHERE clause
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if manufacturer != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("manufacturer ILIKE $%d", argIndex))
		args = append(args, "%"+manufacturer+"%")
		argIndex++
	}

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if batchID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("batch_id = $%d", argIndex))
		args = append(args, batchID)
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM aluminium_passports WHERE %s", whereClause)
	var total int
	err := db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT id, passport_id, batch_id, manufacturer, origin, status, esg_score, 
		       recycled_content_percent, created_at, updated_at
		FROM aluminium_passports 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var passports []*db.AluminiumPassport
	for rows.Next() {
		passport := &db.AluminiumPassport{}
		err := rows.Scan(
			&passport.ID, &passport.PassportID, &passport.BatchID, &passport.Manufacturer, &passport.Origin,
			&passport.Status, &passport.ESGScore, &passport.RecycledContentPercent,
			&passport.CreatedAt, &passport.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		passports = append(passports, passport)
	}

	return passports, total, nil
}

func (pc *PassportController) addSupplyChainStep(passportID, stepName, description string, userID int) error {
	query := `
		INSERT INTO supply_chain_steps (passport_id, step_name, step_description, step_date, verifier_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.DB.Exec(query, passportID, stepName, description, time.Now(), userID, time.Now())
	return err
}

// Helper methods
func (pc *PassportController) extractUserClaims(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return auth.ValidateToken(tokenString)
}

func (pc *PassportController) hasRole(userRole string, allowedRoles []string) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func (pc *PassportController) logAuditEvent(userID int, userRole, action, resourceType, resourceID string, oldValues, newValues interface{}, r *http.Request) {
	// Implementation would log to audit_logs table
}

func getIntValue(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func getFloatValue(ptr *float64, defaultValue float64) float64 {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
