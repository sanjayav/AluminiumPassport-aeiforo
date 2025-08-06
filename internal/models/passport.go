package models

import "time"

type AluminiumPassport struct {
	PassportID             string    `json:"passport_id" db:"passport_id"`
	BatchID                string    `json:"batch_id" db:"batch_id"`
	BauxiteOrigin          string    `json:"bauxite_origin" db:"bauxite_origin"`
	MineOperator           string    `json:"mine_operator" db:"mine_operator"`
	DateOfExtraction       string    `json:"date_of_extraction" db:"date_of_extraction"`
	RefineryLocation       string    `json:"refinery_location" db:"refinery_location"`
	RefinerID              string    `json:"refiner_id" db:"refiner_id"`
	SmeltingEnergySource   string    `json:"smelting_energy_source" db:"smelting_energy_source"`
	CarbonEmissionsPerKg   float64   `json:"carbon_emissions_per_kg" db:"carbon_emissions_per_kg"`
	AlloyComposition       string    `json:"alloy_composition" db:"alloy_composition"`
	TraceMetals            string    `json:"trace_metals" db:"trace_metals"`
	ManufacturerID         string    `json:"manufacturer_id" db:"manufacturer_id"`
	ProcessType            string    `json:"process_type" db:"process_type"`
	ManufacturedProduct    string    `json:"manufactured_product" db:"manufactured_product"`
	ProductWeight          float64   `json:"product_weight" db:"product_weight"`
	EnergyUsed             float64   `json:"energy_used" db:"energy_used"`
	ManufacturingEmissions float64   `json:"manufacturing_emissions" db:"manufacturing_emissions"`
	TransportMode          string    `json:"transport_mode" db:"transport_mode"`
	DistanceTravelled      float64   `json:"distance_travelled" db:"distance_travelled"`
	LogisticsPartnerID     string    `json:"logistics_partner_id" db:"logistics_partner_id"`
	ShipmentDate           string    `json:"shipment_date" db:"shipment_date"`
	RecycledContentPercent float64   `json:"recycled_content_percent" db:"recycled_content_percent"`
	RecyclingDate          string    `json:"recycling_date" db:"recycling_date"`
	RecyclerID             string    `json:"recycler_id" db:"recycler_id"`
	RecyclingMethod        string    `json:"recycling_method" db:"recycling_method"`
	TimesRecycled          int       `json:"times_recycled" db:"times_recycled"`
	CertificationAgency    string    `json:"certification_agency" db:"certification_agency"`
	ESGScore               float64   `json:"esg_score" db:"esg_score"`
	ComplianceStandards    string    `json:"compliance_standards" db:"compliance_standards"`
	DateOfCertification    string    `json:"date_of_certification" db:"date_of_certification"`
	VerifierSignature      string    `json:"verifier_signature" db:"verifier_signature"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BatchUploadRequest struct {
	BatchID     string              `json:"batch_id"`
	Passports   []AluminiumPassport `json:"passports"`
	Description string              `json:"description,omitempty"`
}

type BatchUploadResponse struct {
	BatchID        string   `json:"batch_id"`
	TotalProcessed int      `json:"total_processed"`
	Successful     int      `json:"successful"`
	Failed         int      `json:"failed"`
	Errors         []string `json:"errors,omitempty"`
	IPFSHash       string   `json:"ipfs_hash,omitempty"`
}

type FileUploadMetadata struct {
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	UploadedAt  string `json:"uploaded_at"`
	ProcessedBy string `json:"processed_by"`
}
