package qr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"

	"aluminium-passport/internal/config"
	"aluminium-passport/internal/db"

	"github.com/skip2/go-qrcode"
)

type QRCodeData struct {
	PassportID          string   `json:"passport_id"`
	Manufacturer        string   `json:"manufacturer"`
	Origin              string   `json:"origin"`
	RecycledContent     *float64 `json:"recycled_content,omitempty"`
	ESGScore            *float64 `json:"esg_score,omitempty"`
	CertificationAgency *string  `json:"certification_agency,omitempty"`
	IPFSHash            *string  `json:"ipfs_hash,omitempty"`
	VerificationURL     string   `json:"verification_url"`
	GeneratedAt         string   `json:"generated_at"`
	Version             string   `json:"version"`
}

const (
	QRCodeSize      = 256
	QRRecoveryLevel = qrcode.Medium
	QRVersion       = "1.0"
)

// GeneratePassportQR generates QR code data (JSON string) for a passport
func GeneratePassportQR(passport *db.AluminiumPassport) (string, error) {
	cfg := config.AppConfig

	// Create QR code data structure
	qrData := &QRCodeData{
		PassportID:          passport.PassportID,
		Manufacturer:        passport.Manufacturer,
		Origin:              passport.Origin,
		RecycledContent:     passport.RecycledContentPercent,
		ESGScore:            passport.ESGScore,
		CertificationAgency: passport.CertificationAgency,
		IPFSHash:            passport.IPFSHash,
		VerificationURL:     fmt.Sprintf("%s/verify/%s", getBaseURL(cfg), passport.PassportID),
		GeneratedAt:         passport.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Version:             QRVersion,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal QR data: %w", err)
	}

	return string(jsonData), nil
}

// GenerateQRCodeImage generates a QR code image (PNG) for a passport
func GenerateQRCodeImage(passport *db.AluminiumPassport) ([]byte, error) {
	// Get QR data
	qrDataJSON, err := GeneratePassportQR(passport)
	if err != nil {
		return nil, err
	}

	// Generate QR code
	qr, err := qrcode.New(qrDataJSON, QRRecoveryLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	// Convert to PNG
	img := qr.Image(QRCodeSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code as PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateQRCodeBase64 generates a base64-encoded QR code image
func GenerateQRCodeBase64(passport *db.AluminiumPassport) (string, error) {
	imageData, err := GenerateQRCodeImage(passport)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}

// GenerateGS1QRCode generates a GS1-compliant QR code for supply chain tracking
func GenerateGS1QRCode(passport *db.AluminiumPassport) ([]byte, error) {
	// GS1 format: (01)GTIN(21)SerialNumber(10)BatchLot(17)ExpiryDate
	// For aluminium passport, we'll use a simplified format

	gs1Data := fmt.Sprintf("(01)%s(21)%s(10)%s",
		generateGTIN(passport.Manufacturer, passport.PassportID),
		passport.PassportID,
		getStringValue(passport.BatchID, "NOBATCH"))

	// Add recycled content if available
	if passport.RecycledContentPercent != nil {
		gs1Data += fmt.Sprintf("(7003)%.0f", *passport.RecycledContentPercent*100) // Weight percentage
	}

	// Add ESG score if available
	if passport.ESGScore != nil {
		gs1Data += fmt.Sprintf("(7004)%.0f", *passport.ESGScore*100) // ESG score as percentage
	}

	// Generate QR code
	qr, err := qrcode.New(gs1Data, QRRecoveryLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create GS1 QR code: %w", err)
	}

	// Convert to PNG
	img := qr.Image(QRCodeSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode GS1 QR code as PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// ParseQRCode parses QR code data and returns the passport information
func ParseQRCode(qrData string) (*QRCodeData, error) {
	var data QRCodeData
	if err := json.Unmarshal([]byte(qrData), &data); err != nil {
		return nil, fmt.Errorf("failed to parse QR code data: %w", err)
	}

	// Validate required fields
	if data.PassportID == "" {
		return nil, fmt.Errorf("invalid QR code: missing passport ID")
	}

	if data.Version != QRVersion {
		return nil, fmt.Errorf("unsupported QR code version: %s", data.Version)
	}

	return &data, nil
}

// ValidateQRCode validates a QR code against a passport
func ValidateQRCode(qrData string, passport *db.AluminiumPassport) error {
	parsedData, err := ParseQRCode(qrData)
	if err != nil {
		return err
	}

	// Validate passport ID matches
	if parsedData.PassportID != passport.PassportID {
		return fmt.Errorf("QR code passport ID mismatch")
	}

	// Validate manufacturer matches
	if parsedData.Manufacturer != passport.Manufacturer {
		return fmt.Errorf("QR code manufacturer mismatch")
	}

	// Validate origin matches
	if parsedData.Origin != passport.Origin {
		return fmt.Errorf("QR code origin mismatch")
	}

	return nil
}

// GenerateBatchQRCode generates a QR code for a batch of passports
func GenerateBatchQRCode(batchID string, passportCount int, ipfsHash *string) ([]byte, error) {
	cfg := config.AppConfig

	batchData := map[string]interface{}{
		"batch_id":         batchID,
		"passport_count":   passportCount,
		"ipfs_hash":        ipfsHash,
		"verification_url": fmt.Sprintf("%s/verify/batch/%s", getBaseURL(cfg), batchID),
		"type":             "batch",
		"version":          QRVersion,
	}

	jsonData, err := json.Marshal(batchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch QR data: %w", err)
	}

	// Generate QR code
	qr, err := qrcode.New(string(jsonData), QRRecoveryLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch QR code: %w", err)
	}

	// Convert to PNG
	img := qr.Image(QRCodeSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode batch QR code as PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateESGQRCode generates a QR code specifically for ESG verification
func GenerateESGQRCode(passport *db.AluminiumPassport, esgMetrics *db.ESGMetrics) ([]byte, error) {
	cfg := config.AppConfig

	esgData := map[string]interface{}{
		"passport_id":         passport.PassportID,
		"manufacturer":        passport.Manufacturer,
		"overall_esg_score":   esgMetrics.OverallESGScore,
		"environmental_score": esgMetrics.EnergyEfficiencyScore,
		"social_score":        esgMetrics.LaborPracticesScore,
		"governance_score":    esgMetrics.ComplianceScore,
		"assessment_date":     esgMetrics.AssessmentDate,
		"verification_url":    fmt.Sprintf("%s/verify/esg/%s", getBaseURL(cfg), passport.PassportID),
		"type":                "esg_verification",
		"version":             QRVersion,
	}

	jsonData, err := json.Marshal(esgData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ESG QR data: %w", err)
	}

	// Generate QR code
	qr, err := qrcode.New(string(jsonData), QRRecoveryLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create ESG QR code: %w", err)
	}

	// Convert to PNG
	img := qr.Image(QRCodeSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode ESG QR code as PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// DecodeQRCodeFromImage decodes QR code from an image
func DecodeQRCodeFromImage(img image.Image) (string, error) {
	// This would require a QR code decoder library
	// For now, return a placeholder implementation
	return "", fmt.Errorf("QR code decoding from image not implemented")
}

// Helper functions
func generateGTIN(manufacturer, passportID string) string {
	// Generate a simplified GTIN-like identifier
	// In production, this would follow actual GTIN standards
	return fmt.Sprintf("0%s%s",
		fmt.Sprintf("%06d", hashString(manufacturer)%1000000),
		fmt.Sprintf("%06d", hashString(passportID)%1000000))
}

func hashString(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func getStringValue(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func getBaseURL(cfg *config.Config) string {
	// In production, this should come from configuration
	if cfg.Environment == "production" {
		return "https://aluminium-passport.com"
	}
	return fmt.Sprintf("http://localhost:%s", cfg.Port)
}
