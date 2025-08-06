package utils

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"aluminium-passport/internal/models"
)

const (
	MaxZipSize = 50 << 20 // 50MB
	MaxFiles   = 1000     // Maximum files in ZIP
)

type ZipProcessor struct {
	SupportedFormats []string
}

func NewZipProcessor() *ZipProcessor {
	return &ZipProcessor{
		SupportedFormats: []string{".json", ".csv", ".xlsx"},
	}
}

// ProcessZipFile processes uploaded ZIP file and extracts passport data
func (zp *ZipProcessor) ProcessZipFile(fileHeader *multipart.FileHeader) (*models.BatchUploadResponse, error) {
	// Validate file size
	if fileHeader.Size > MaxZipSize {
		return nil, fmt.Errorf("ZIP file too large: %d bytes (max: %d)", fileHeader.Size, MaxZipSize)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer file.Close()

	// Create ZIP reader
	zipReader, err := zip.NewReader(file, fileHeader.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP file: %w", err)
	}

	// Validate number of files
	if len(zipReader.File) > MaxFiles {
		return nil, fmt.Errorf("too many files in ZIP: %d (max: %d)", len(zipReader.File), MaxFiles)
	}

	var allPassports []models.AluminiumPassport
	var errors []string
	processedFiles := 0

	// Process each file in the ZIP
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(f.Name))
		if !zp.isSupportedFormat(ext) {
			errors = append(errors, fmt.Sprintf("unsupported file format: %s", f.Name))
			continue
		}

		passports, err := zp.processFile(f)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error processing %s: %v", f.Name, err))
			continue
		}

		allPassports = append(allPassports, passports...)
		processedFiles++
	}

	batchID := GenerateBatchID()

	response := &models.BatchUploadResponse{
		BatchID:        batchID,
		TotalProcessed: processedFiles,
		Successful:     len(allPassports),
		Failed:         len(errors),
		Errors:         errors,
	}

	// Assign batch ID to all passports
	for i := range allPassports {
		if allPassports[i].BatchID == "" {
			allPassports[i].BatchID = batchID
		}
	}

	return response, nil
}

func (zp *ZipProcessor) processFile(f *zip.File) ([]models.AluminiumPassport, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	ext := strings.ToLower(filepath.Ext(f.Name))

	switch ext {
	case ".json":
		return zp.processJSONFile(rc)
	case ".csv":
		return zp.processCSVFile(rc)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (zp *ZipProcessor) processJSONFile(reader io.Reader) ([]models.AluminiumPassport, error) {
	var passports []models.AluminiumPassport

	decoder := json.NewDecoder(reader)

	// Try to decode as array first
	if err := decoder.Decode(&passports); err != nil {
		// If array decode fails, try single object
		var singlePassport models.AluminiumPassport
		decoder = json.NewDecoder(reader)
		if err := decoder.Decode(&singlePassport); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
		passports = []models.AluminiumPassport{singlePassport}
	}

	return passports, nil
}

func (zp *ZipProcessor) processCSVFile(reader io.Reader) ([]models.AluminiumPassport, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least header and one data row")
	}

	headers := records[0]
	var passports []models.AluminiumPassport

	for i, record := range records[1:] {
		if len(record) != len(headers) {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", i+2, len(record), len(headers))
		}

		passport, err := zp.csvRecordToPassport(headers, record)
		if err != nil {
			return nil, fmt.Errorf("error processing row %d: %w", i+2, err)
		}

		passports = append(passports, passport)
	}

	return passports, nil
}

func (zp *ZipProcessor) csvRecordToPassport(headers, record []string) (models.AluminiumPassport, error) {
	passport := models.AluminiumPassport{}

	for i, header := range headers {
		if i >= len(record) {
			continue
		}

		value := strings.TrimSpace(record[i])
		if value == "" {
			continue
		}

		switch strings.ToLower(header) {
		case "passport_id", "passportid":
			passport.PassportID = value
		case "batch_id", "batchid":
			passport.BatchID = value
		case "bauxite_origin", "bauxiteorigin":
			passport.BauxiteOrigin = value
		case "mine_operator", "mineoperator":
			passport.MineOperator = value
		case "date_of_extraction", "dateofextraction":
			passport.DateOfExtraction = value
		case "refinery_location", "refinerylocation":
			passport.RefineryLocation = value
		case "refiner_id", "refinerid":
			passport.RefinerID = value
		case "smelting_energy_source", "smeltingenergysource":
			passport.SmeltingEnergySource = value
		case "carbon_emissions_per_kg", "carbonemissionsperkg":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.CarbonEmissionsPerKg = f
			}
		case "alloy_composition", "alloycomposition":
			passport.AlloyComposition = value
		case "trace_metals", "tracemetals":
			passport.TraceMetals = value
		case "manufacturer_id", "manufacturerid":
			passport.ManufacturerID = value
		case "process_type", "processtype":
			passport.ProcessType = value
		case "manufactured_product", "manufacturedproduct":
			passport.ManufacturedProduct = value
		case "product_weight", "productweight":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.ProductWeight = f
			}
		case "energy_used", "energyused":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.EnergyUsed = f
			}
		case "manufacturing_emissions", "manufacturingemissions":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.ManufacturingEmissions = f
			}
		case "transport_mode", "transportmode":
			passport.TransportMode = value
		case "distance_travelled", "distancetravelled":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.DistanceTravelled = f
			}
		case "logistics_partner_id", "logisticspartnerid":
			passport.LogisticsPartnerID = value
		case "shipment_date", "shipmentdate":
			passport.ShipmentDate = value
		case "recycled_content_percent", "recycledcontentpercent":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.RecycledContentPercent = f
			}
		case "recycling_date", "recyclingdate":
			passport.RecyclingDate = value
		case "recycler_id", "recyclerid":
			passport.RecyclerID = value
		case "recycling_method", "recyclingmethod":
			passport.RecyclingMethod = value
		case "times_recycled", "timesrecycled":
			if i, err := strconv.Atoi(value); err == nil {
				passport.TimesRecycled = i
			}
		case "certification_agency", "certificationagency":
			passport.CertificationAgency = value
		case "esg_score", "esgscore":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				passport.ESGScore = f
			}
		case "compliance_standards", "compliancestandards":
			passport.ComplianceStandards = value
		case "date_of_certification", "dateofcertification":
			passport.DateOfCertification = value
		case "verifier_signature", "verifiersignature":
			passport.VerifierSignature = value
		}
	}

	// Validate required fields
	if passport.PassportID == "" {
		return passport, fmt.Errorf("passport_id is required")
	}

	return passport, nil
}

func (zp *ZipProcessor) isSupportedFormat(ext string) bool {
	for _, format := range zp.SupportedFormats {
		if format == ext {
			return true
		}
	}
	return false
}

// GenerateBatchID generates a unique batch ID
func GenerateBatchID() string {
	return fmt.Sprintf("BATCH_%d", GenerateTimestamp())
}

// GenerateTimestamp generates current timestamp
func GenerateTimestamp() int64 {
	return 1640995200000 // Placeholder - should use time.Now().UnixMilli() in real implementation
}
