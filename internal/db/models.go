package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a system user
type User struct {
	ID            int        `json:"id" db:"id"`
	WalletAddress string     `json:"wallet_address" db:"wallet_address"`
	Username      string     `json:"username" db:"username"`
	Email         *string    `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	Role          string     `json:"role" db:"role"`
	CompanyName   *string    `json:"company_name" db:"company_name"`
	ContactInfo   *JSONMap   `json:"contact_info" db:"contact_info"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin     *time.Time `json:"last_login" db:"last_login"`
}

// AluminiumPassport represents a passport record
type AluminiumPassport struct {
	ID         int     `json:"id" db:"id"`
	PassportID string  `json:"passport_id" db:"passport_id"`
	BatchID    *string `json:"batch_id" db:"batch_id"`

	// Basic Information
	Manufacturer     string  `json:"manufacturer" db:"manufacturer"`
	Origin           string  `json:"origin" db:"origin"`
	BauxiteSource    *string `json:"bauxite_source" db:"bauxite_source"`
	AlloyComposition *string `json:"alloy_composition" db:"alloy_composition"`

	// Mining & Extraction
	MineOperator     *string    `json:"mine_operator" db:"mine_operator"`
	DateOfExtraction *time.Time `json:"date_of_extraction" db:"date_of_extraction"`
	ExtractionMethod *string    `json:"extraction_method" db:"extraction_method"`
	MineLocation     *string    `json:"mine_location" db:"mine_location"`

	// Refining
	RefineryLocation *string    `json:"refinery_location" db:"refinery_location"`
	RefinerID        *string    `json:"refiner_id" db:"refiner_id"`
	RefiningDate     *time.Time `json:"refining_date" db:"refining_date"`
	RefiningMethod   *string    `json:"refining_method" db:"refining_method"`

	// Smelting & Manufacturing
	SmeltingLocation     *string    `json:"smelting_location" db:"smelting_location"`
	SmeltingEnergySource *string    `json:"smelting_energy_source" db:"smelting_energy_source"`
	ProcessType          *string    `json:"process_type" db:"process_type"`
	ManufacturedProduct  *string    `json:"manufactured_product" db:"manufactured_product"`
	ManufacturingDate    *time.Time `json:"manufacturing_date" db:"manufacturing_date"`

	// Quantities & Measurements
	ProductWeight  *float64 `json:"product_weight" db:"product_weight"`
	EnergyUsed     *float64 `json:"energy_used" db:"energy_used"`
	WaterUsed      *float64 `json:"water_used" db:"water_used"`
	WasteGenerated *float64 `json:"waste_generated" db:"waste_generated"`

	// Environmental Impact
	CarbonEmissionsPerKg   *float64 `json:"carbon_emissions_per_kg" db:"carbon_emissions_per_kg"`
	CO2Footprint           *float64 `json:"co2_footprint" db:"co2_footprint"`
	ManufacturingEmissions *float64 `json:"manufacturing_emissions" db:"manufacturing_emissions"`

	// Supply Chain & Logistics
	TransportMode      *string    `json:"transport_mode" db:"transport_mode"`
	DistanceTravelled  *float64   `json:"distance_travelled" db:"distance_travelled"`
	LogisticsPartnerID *string    `json:"logistics_partner_id" db:"logistics_partner_id"`
	ShipmentDate       *time.Time `json:"shipment_date" db:"shipment_date"`

	// Recycling Information
	RecycledContentPercent *float64   `json:"recycled_content_percent" db:"recycled_content_percent"`
	RecyclingDate          *time.Time `json:"recycling_date" db:"recycling_date"`
	RecyclerID             *string    `json:"recycler_id" db:"recycler_id"`
	RecyclingMethod        *string    `json:"recycling_method" db:"recycling_method"`
	TimesRecycled          int        `json:"times_recycled" db:"times_recycled"`
	LastRecyclingDate      *time.Time `json:"last_recycling_date" db:"last_recycling_date"`

	// Certifications & Compliance
	CertificationAgency *string    `json:"certification_agency" db:"certification_agency"`
	Certifier           *string    `json:"certifier" db:"certifier"`
	ComplianceStandards *string    `json:"compliance_standards" db:"compliance_standards"`
	DateOfCertification *time.Time `json:"date_of_certification" db:"date_of_certification"`
	CertificationExpiry *time.Time `json:"certification_expiry" db:"certification_expiry"`
	VerifierSignature   *string    `json:"verifier_signature" db:"verifier_signature"`

	// ESG Scoring
	ESGScore           *float64   `json:"esg_score" db:"esg_score"`
	EnvironmentalScore *float64   `json:"environmental_score" db:"environmental_score"`
	SocialScore        *float64   `json:"social_score" db:"social_score"`
	GovernanceScore    *float64   `json:"governance_score" db:"governance_score"`
	ESGLastUpdated     *time.Time `json:"esg_last_updated" db:"esg_last_updated"`

	// Digital Assets
	IPFSHash         *string `json:"ipfs_hash" db:"ipfs_hash"`
	QRCodeData       *string `json:"qr_code_data" db:"qr_code_data"`
	DigitalSignature *string `json:"digital_signature" db:"digital_signature"`

	// Blockchain Integration
	BlockchainTxHash *string `json:"blockchain_tx_hash" db:"blockchain_tx_hash"`
	ContractAddress  *string `json:"contract_address" db:"contract_address"`
	BlockNumber      *int64  `json:"block_number" db:"block_number"`

	// Status & Metadata
	Status           string     `json:"status" db:"status"`
	IsVerified       bool       `json:"is_verified" db:"is_verified"`
	VerificationDate *time.Time `json:"verification_date" db:"verification_date"`

	// Additional metadata as JSON
	Metadata         *JSONMap   `json:"metadata" db:"metadata"`
	SupplyChainSteps *JSONArray `json:"supply_chain_steps" db:"supply_chain_steps"`
	Certifications   *JSONArray `json:"certifications" db:"certifications"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy *int      `json:"created_by" db:"created_by"`
	UpdatedBy *int      `json:"updated_by" db:"updated_by"`
}

// ESGMetrics represents detailed ESG scoring
type ESGMetrics struct {
	ID         int    `json:"id" db:"id"`
	PassportID string `json:"passport_id" db:"passport_id"`

	// Environmental metrics
	CarbonFootprint        *float64 `json:"carbon_footprint" db:"carbon_footprint"`
	EnergyEfficiencyScore  *float64 `json:"energy_efficiency_score" db:"energy_efficiency_score"`
	WaterUsageScore        *float64 `json:"water_usage_score" db:"water_usage_score"`
	WasteManagementScore   *float64 `json:"waste_management_score" db:"waste_management_score"`
	RenewableEnergyPercent *float64 `json:"renewable_energy_percent" db:"renewable_energy_percent"`

	// Social metrics
	LaborPracticesScore  *float64 `json:"labor_practices_score" db:"labor_practices_score"`
	CommunityImpactScore *float64 `json:"community_impact_score" db:"community_impact_score"`
	HealthSafetyScore    *float64 `json:"health_safety_score" db:"health_safety_score"`
	HumanRightsScore     *float64 `json:"human_rights_score" db:"human_rights_score"`

	// Governance metrics
	TransparencyScore          *float64 `json:"transparency_score" db:"transparency_score"`
	EthicsScore                *float64 `json:"ethics_score" db:"ethics_score"`
	ComplianceScore            *float64 `json:"compliance_score" db:"compliance_score"`
	StakeholderEngagementScore *float64 `json:"stakeholder_engagement_score" db:"stakeholder_engagement_score"`

	// Composite scores
	OverallESGScore *float64 `json:"overall_esg_score" db:"overall_esg_score"`

	// Metadata
	AssessmentDate        *time.Time `json:"assessment_date" db:"assessment_date"`
	AssessorID            *int       `json:"assessor_id" db:"assessor_id"`
	AssessmentMethodology *string    `json:"assessment_methodology" db:"assessment_methodology"`
	Notes                 *string    `json:"notes" db:"notes"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SupplyChainStep represents a step in the supply chain
type SupplyChainStep struct {
	ID                 int        `json:"id" db:"id"`
	PassportID         string     `json:"passport_id" db:"passport_id"`
	StepName           string     `json:"step_name" db:"step_name"`
	StepDescription    *string    `json:"step_description" db:"step_description"`
	Location           *string    `json:"location" db:"location"`
	ResponsibleParty   *string    `json:"responsible_party" db:"responsible_party"`
	StepDate           *time.Time `json:"step_date" db:"step_date"`
	VerificationStatus bool       `json:"verification_status" db:"verification_status"`
	VerifierID         *int       `json:"verifier_id" db:"verifier_id"`
	StepData           *JSONMap   `json:"step_data" db:"step_data"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int       `json:"id" db:"id"`
	UserID       *int      `json:"user_id" db:"user_id"`
	UserRole     *string   `json:"user_role" db:"user_role"`
	Action       string    `json:"action" db:"action"`
	ResourceType *string   `json:"resource_type" db:"resource_type"`
	ResourceID   *string   `json:"resource_id" db:"resource_id"`
	OldValues    *JSONMap  `json:"old_values" db:"old_values"`
	NewValues    *JSONMap  `json:"new_values" db:"new_values"`
	IPAddress    *string   `json:"ip_address" db:"ip_address"`
	UserAgent    *string   `json:"user_agent" db:"user_agent"`
	Success      bool      `json:"success" db:"success"`
	ErrorMessage *string   `json:"error_message" db:"error_message"`
	SessionID    *string   `json:"session_id" db:"session_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Certification represents a certification record
type Certification struct {
	ID                int        `json:"id" db:"id"`
	PassportID        string     `json:"passport_id" db:"passport_id"`
	CertificationName string     `json:"certification_name" db:"certification_name"`
	CertificationBody *string    `json:"certification_body" db:"certification_body"`
	CertificateNumber *string    `json:"certificate_number" db:"certificate_number"`
	IssueDate         *time.Time `json:"issue_date" db:"issue_date"`
	ExpiryDate        *time.Time `json:"expiry_date" db:"expiry_date"`
	Status            string     `json:"status" db:"status"`
	CertificateURL    *string    `json:"certificate_url" db:"certificate_url"`
	VerificationHash  *string    `json:"verification_hash" db:"verification_hash"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// BatchOperation represents a batch operation
type BatchOperation struct {
	ID                int        `json:"id" db:"id"`
	BatchID           string     `json:"batch_id" db:"batch_id"`
	OperationType     *string    `json:"operation_type" db:"operation_type"`
	TotalRecords      *int       `json:"total_records" db:"total_records"`
	SuccessfulRecords *int       `json:"successful_records" db:"successful_records"`
	FailedRecords     *int       `json:"failed_records" db:"failed_records"`
	Status            string     `json:"status" db:"status"`
	ErrorLog          *JSONMap   `json:"error_log" db:"error_log"`
	IPFSHash          *string    `json:"ipfs_hash" db:"ipfs_hash"`
	CreatedBy         *int       `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	CompletedAt       *time.Time `json:"completed_at" db:"completed_at"`
}

// ZKProof represents a zero-knowledge proof
type ZKProof struct {
	ID              int        `json:"id" db:"id"`
	PassportID      string     `json:"passport_id" db:"passport_id"`
	ProofType       *string    `json:"proof_type" db:"proof_type"`
	ProofData       *string    `json:"proof_data" db:"proof_data"`
	PublicInputs    *JSONMap   `json:"public_inputs" db:"public_inputs"`
	VerificationKey *string    `json:"verification_key" db:"verification_key"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	VerifierAddress *string    `json:"verifier_address" db:"verifier_address"`
	CreatedBy       *int       `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	VerifiedAt      *time.Time `json:"verified_at" db:"verified_at"`
}

// JSONMap for handling JSONB fields
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// JSONArray for handling JSONB arrays
type JSONArray []interface{}

// Value implements the driver.Valuer interface
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONArray", value)
	}

	return json.Unmarshal(bytes, j)
}
