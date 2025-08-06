package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"aluminium-passport/internal/config"
	"aluminium-passport/internal/db"
)

type IPFSClient struct {
	apiURL        string
	projectID     string
	projectSecret string
	gatewayURL    string
	httpClient    *http.Client
}

type IPFSResponse struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

type PassportIPFSData struct {
	PassportID             string        `json:"passport_id"`
	Manufacturer           string        `json:"manufacturer"`
	Origin                 string        `json:"origin"`
	AlloyComposition       *string       `json:"alloy_composition,omitempty"`
	RecycledContentPercent *float64      `json:"recycled_content_percent,omitempty"`
	ESGScore               *float64      `json:"esg_score,omitempty"`
	CarbonEmissionsPerKg   *float64      `json:"carbon_emissions_per_kg,omitempty"`
	CertificationAgency    *string       `json:"certification_agency,omitempty"`
	SupplyChainSteps       *db.JSONArray `json:"supply_chain_steps,omitempty"`
	Certifications         *db.JSONArray `json:"certifications,omitempty"`
	Metadata               *db.JSONMap   `json:"metadata,omitempty"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
	Version                string        `json:"version"`
	IPFSTimestamp          time.Time     `json:"ipfs_timestamp"`
}

var DefaultClient *IPFSClient

func InitializeIPFS() error {
	cfg := config.AppConfig

	DefaultClient = &IPFSClient{
		apiURL:        cfg.IPFSAPIUrl,
		projectID:     cfg.IPFSProjectID,
		projectSecret: cfg.IPFSProjectSecret,
		gatewayURL:    cfg.IPFSGatewayURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Test connection
	return DefaultClient.TestConnection()
}

func (c *IPFSClient) TestConnection() error {
	// Simple test to check if IPFS service is accessible
	req, err := http.NewRequest("POST", c.apiURL+"/api/v0/version", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	if c.projectID != "" && c.projectSecret != "" {
		req.SetBasicAuth(c.projectID, c.projectSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPFS service returned status: %d", resp.StatusCode)
	}

	return nil
}

func (c *IPFSClient) UploadJSON(data interface{}) (string, error) {
	// Marshal data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fw, err := w.CreateFormFile("file", "data.json")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = fw.Write(jsonData); err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	w.Close()

	// Create request
	req, err := http.NewRequest("POST", c.apiURL+"/api/v0/add", &b)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	if c.projectID != "" && c.projectSecret != "" {
		req.SetBasicAuth(c.projectID, c.projectSecret)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("IPFS upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result IPFSResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse IPFS response: %w", err)
	}

	return "ipfs://" + result.Hash, nil
}

func (c *IPFSClient) RetrieveJSON(hash string, target interface{}) error {
	// Remove ipfs:// prefix if present
	hash = strings.TrimPrefix(hash, "ipfs://")

	// Build gateway URL
	url := c.gatewayURL + hash

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to retrieve from IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPFS retrieval failed with status: %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (c *IPFSClient) Pin(hash string) error {
	// Remove ipfs:// prefix if present
	hash = strings.TrimPrefix(hash, "ipfs://")

	// Create request
	req, err := http.NewRequest("POST", c.apiURL+"/api/v0/pin/add?arg="+hash, nil)
	if err != nil {
		return fmt.Errorf("failed to create pin request: %w", err)
	}

	if c.projectID != "" && c.projectSecret != "" {
		req.SetBasicAuth(c.projectID, c.projectSecret)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pin to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IPFS pin failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *IPFSClient) Unpin(hash string) error {
	// Remove ipfs:// prefix if present
	hash = strings.TrimPrefix(hash, "ipfs://")

	// Create request
	req, err := http.NewRequest("POST", c.apiURL+"/api/v0/pin/rm?arg="+hash, nil)
	if err != nil {
		return fmt.Errorf("failed to create unpin request: %w", err)
	}

	if c.projectID != "" && c.projectSecret != "" {
		req.SetBasicAuth(c.projectID, c.projectSecret)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unpin from IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IPFS unpin failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// High-level functions for passport data
func UploadPassportData(passport *db.AluminiumPassport) (string, error) {
	if DefaultClient == nil {
		return "", fmt.Errorf("IPFS client not initialized")
	}

	// Create IPFS-specific data structure
	ipfsData := &PassportIPFSData{
		PassportID:             passport.PassportID,
		Manufacturer:           passport.Manufacturer,
		Origin:                 passport.Origin,
		AlloyComposition:       passport.AlloyComposition,
		RecycledContentPercent: passport.RecycledContentPercent,
		ESGScore:               passport.ESGScore,
		CarbonEmissionsPerKg:   passport.CarbonEmissionsPerKg,
		CertificationAgency:    passport.CertificationAgency,
		SupplyChainSteps:       passport.SupplyChainSteps,
		Certifications:         passport.Certifications,
		Metadata:               passport.Metadata,
		CreatedAt:              passport.CreatedAt,
		UpdatedAt:              passport.UpdatedAt,
		Version:                "1.0",
		IPFSTimestamp:          time.Now(),
	}

	hash, err := DefaultClient.UploadJSON(ipfsData)
	if err != nil {
		return "", err
	}

	// Pin the content to ensure it stays available
	if err := DefaultClient.Pin(hash); err != nil {
		// Log warning but don't fail the upload
		fmt.Printf("Warning: Failed to pin IPFS content %s: %v\n", hash, err)
	}

	return hash, nil
}

func RetrievePassportData(hash string) (*PassportIPFSData, error) {
	if DefaultClient == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	var data PassportIPFSData
	err := DefaultClient.RetrieveJSON(hash, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func UploadBatchData(batchID string, passports []*db.AluminiumPassport, metadata map[string]interface{}) (string, error) {
	if DefaultClient == nil {
		return "", fmt.Errorf("IPFS client not initialized")
	}

	// Create batch data structure
	batchData := map[string]interface{}{
		"batch_id":    batchID,
		"passports":   passports,
		"metadata":    metadata,
		"created_at":  time.Now(),
		"version":     "1.0",
		"total_count": len(passports),
	}

	hash, err := DefaultClient.UploadJSON(batchData)
	if err != nil {
		return "", err
	}

	// Pin the batch data
	if err := DefaultClient.Pin(hash); err != nil {
		fmt.Printf("Warning: Failed to pin batch IPFS content %s: %v\n", hash, err)
	}

	return hash, nil
}

func UploadESGReport(passportID string, esgMetrics *db.ESGMetrics) (string, error) {
	if DefaultClient == nil {
		return "", fmt.Errorf("IPFS client not initialized")
	}

	// Create ESG report structure
	esgReport := map[string]interface{}{
		"passport_id":      passportID,
		"esg_metrics":      esgMetrics,
		"report_generated": time.Now(),
		"version":          "1.0",
		"report_type":      "esg_assessment",
	}

	return DefaultClient.UploadJSON(esgReport)
}

func GetIPFSGatewayURL(hash string) string {
	if DefaultClient == nil {
		return ""
	}

	hash = strings.TrimPrefix(hash, "ipfs://")
	return DefaultClient.gatewayURL + hash
}

func IsIPFSAvailable() bool {
	if DefaultClient == nil {
		return false
	}

	return DefaultClient.TestConnection() == nil
}
