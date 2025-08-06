package controller

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aluminium-passport/internal/auth"
	"aluminium-passport/internal/db"

	"github.com/gorilla/mux"
)

type ESGController struct{}

func NewESGController() *ESGController {
	return &ESGController{}
}

type ESGAssessmentRequest struct {
	PassportID                 string   `json:"passport_id" binding:"required"`
	CarbonFootprint            *float64 `json:"carbon_footprint"`
	EnergyEfficiencyScore      *float64 `json:"energy_efficiency_score"`
	WaterUsageScore            *float64 `json:"water_usage_score"`
	WasteManagementScore       *float64 `json:"waste_management_score"`
	RenewableEnergyPercent     *float64 `json:"renewable_energy_percent"`
	LaborPracticesScore        *float64 `json:"labor_practices_score"`
	CommunityImpactScore       *float64 `json:"community_impact_score"`
	HealthSafetyScore          *float64 `json:"health_safety_score"`
	HumanRightsScore           *float64 `json:"human_rights_score"`
	TransparencyScore          *float64 `json:"transparency_score"`
	EthicsScore                *float64 `json:"ethics_score"`
	ComplianceScore            *float64 `json:"compliance_score"`
	StakeholderEngagementScore *float64 `json:"stakeholder_engagement_score"`
	AssessmentMethodology      *string  `json:"assessment_methodology"`
	Notes                      *string  `json:"notes"`
}

type ESGScoreRequest struct {
	// Manufacturing data for AI-based scoring
	EnergyUsed               *float64 `json:"energy_used"`
	RenewableEnergyPercent   *float64 `json:"renewable_energy_percent"`
	WaterUsed                *float64 `json:"water_used"`
	WasteGenerated           *float64 `json:"waste_generated"`
	CarbonEmissions          *float64 `json:"carbon_emissions"`
	RecycledContentPercent   *float64 `json:"recycled_content_percent"`
	SupplyChainTransparency  *float64 `json:"supply_chain_transparency"`
	LaborStandardsCompliance *float64 `json:"labor_standards_compliance"`
	CommunityInvestment      *float64 `json:"community_investment"`
	SafetyIncidentRate       *float64 `json:"safety_incident_rate"`
	EthicsTrainingHours      *float64 `json:"ethics_training_hours"`
	BoardDiversity           *float64 `json:"board_diversity"`
	DataPrivacyCompliance    *float64 `json:"data_privacy_compliance"`
}

type ESGResponse struct {
	PassportID            string             `json:"passport_id"`
	OverallESGScore       *float64           `json:"overall_esg_score"`
	EnvironmentalScore    *float64           `json:"environmental_score"`
	SocialScore           *float64           `json:"social_score"`
	GovernanceScore       *float64           `json:"governance_score"`
	DetailedMetrics       *db.ESGMetrics     `json:"detailed_metrics,omitempty"`
	ScoreBreakdown        map[string]float64 `json:"score_breakdown"`
	Recommendations       []string           `json:"recommendations"`
	CertificationLevel    string             `json:"certification_level"`
	LastUpdated           time.Time          `json:"last_updated"`
	AssessmentMethodology *string            `json:"assessment_methodology"`
	Notes                 *string            `json:"notes"`
}

// CreateESGAssessment creates or updates ESG metrics for a passport
func (ec *ESGController) CreateESGAssessment(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	claims, err := ec.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has certifier role
	if !ec.hasRole(claims.Role, []string{"certifier", "admin"}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req ESGAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate passport exists
	if exists, err := ec.passportExists(req.PassportID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if !exists {
		http.Error(w, "Passport not found", http.StatusNotFound)
		return
	}

	// Calculate composite scores
	envScore := ec.calculateEnvironmentalScore(&req)
	socialScore := ec.calculateSocialScore(&req)
	govScore := ec.calculateGovernanceScore(&req)
	overallScore := (envScore + socialScore + govScore) / 3

	// Create ESG metrics
	esgMetrics := &db.ESGMetrics{
		PassportID:                 req.PassportID,
		CarbonFootprint:            req.CarbonFootprint,
		EnergyEfficiencyScore:      req.EnergyEfficiencyScore,
		WaterUsageScore:            req.WaterUsageScore,
		WasteManagementScore:       req.WasteManagementScore,
		RenewableEnergyPercent:     req.RenewableEnergyPercent,
		LaborPracticesScore:        req.LaborPracticesScore,
		CommunityImpactScore:       req.CommunityImpactScore,
		HealthSafetyScore:          req.HealthSafetyScore,
		HumanRightsScore:           req.HumanRightsScore,
		TransparencyScore:          req.TransparencyScore,
		EthicsScore:                req.EthicsScore,
		ComplianceScore:            req.ComplianceScore,
		StakeholderEngagementScore: req.StakeholderEngagementScore,
		OverallESGScore:            &overallScore,
		AssessmentDate:             timePtr(time.Now()),
		AssessorID:                 &claims.UserID,
		AssessmentMethodology:      req.AssessmentMethodology,
		Notes:                      req.Notes,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	// Save to database
	esgID, err := ec.createESGMetrics(esgMetrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create ESG assessment: %v", err), http.StatusInternalServerError)
		return
	}

	esgMetrics.ID = esgID

	// Update passport with ESG scores
	if err := ec.updatePassportESGScores(req.PassportID, envScore, socialScore, govScore, overallScore, claims.UserID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to update passport ESG scores: %v\n", err)
	}

	// Log audit event
	ec.logAuditEvent(claims.UserID, claims.Role, "CREATE", "esg_assessment", req.PassportID, nil, esgMetrics, r)

	// Prepare response
	response := &ESGResponse{
		PassportID:            req.PassportID,
		OverallESGScore:       &overallScore,
		EnvironmentalScore:    &envScore,
		SocialScore:           &socialScore,
		GovernanceScore:       &govScore,
		DetailedMetrics:       esgMetrics,
		ScoreBreakdown:        ec.getScoreBreakdown(esgMetrics),
		Recommendations:       ec.generateRecommendations(esgMetrics),
		CertificationLevel:    ec.getCertificationLevel(overallScore),
		LastUpdated:           time.Now(),
		AssessmentMethodology: req.AssessmentMethodology,
		Notes:                 req.Notes,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetESGMetrics retrieves ESG metrics for a passport
func (ec *ESGController) GetESGMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	passportID := vars["id"]

	// Extract user info from token
	claims, err := ec.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ESG metrics from database
	esgMetrics, err := ec.getESGMetricsByPassportID(passportID)
	if err != nil {
		http.Error(w, "ESG metrics not found", http.StatusNotFound)
		return
	}

	// Calculate composite scores if not stored
	envScore := ec.calculateEnvironmentalScoreFromMetrics(esgMetrics)
	socialScore := ec.calculateSocialScoreFromMetrics(esgMetrics)
	govScore := ec.calculateGovernanceScoreFromMetrics(esgMetrics)

	// Log audit event
	ec.logAuditEvent(claims.UserID, claims.Role, "VIEW", "esg_metrics", passportID, nil, nil, r)

	// Prepare response
	response := &ESGResponse{
		PassportID:            passportID,
		OverallESGScore:       esgMetrics.OverallESGScore,
		EnvironmentalScore:    &envScore,
		SocialScore:           &socialScore,
		GovernanceScore:       &govScore,
		DetailedMetrics:       esgMetrics,
		ScoreBreakdown:        ec.getScoreBreakdown(esgMetrics),
		Recommendations:       ec.generateRecommendations(esgMetrics),
		CertificationLevel:    ec.getCertificationLevel(getFloatValue(esgMetrics.OverallESGScore, 0)),
		LastUpdated:           esgMetrics.UpdatedAt,
		AssessmentMethodology: esgMetrics.AssessmentMethodology,
		Notes:                 esgMetrics.Notes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateESGScore generates AI-based ESG score from manufacturing data
func (ec *ESGController) GenerateESGScore(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	claims, err := ec.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ESGScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate AI-based ESG scores
	envScore := ec.calculateAIEnvironmentalScore(&req)
	socialScore := ec.calculateAISocialScore(&req)
	govScore := ec.calculateAIGovernanceScore(&req)
	overallScore := (envScore + socialScore + govScore) / 3

	// Generate recommendations
	recommendations := ec.generateAIRecommendations(&req, envScore, socialScore, govScore)

	// Log audit event
	ec.logAuditEvent(claims.UserID, claims.Role, "GENERATE", "esg_score", "", nil, req, r)

	response := map[string]interface{}{
		"overall_esg_score":   overallScore,
		"environmental_score": envScore,
		"social_score":        socialScore,
		"governance_score":    govScore,
		"score_breakdown": map[string]float64{
			"energy_efficiency":         getFloatValue(req.RenewableEnergyPercent, 0) * 0.3,
			"waste_management":          (100 - getFloatValue(req.WasteGenerated, 100)) * 0.25,
			"recycled_content":          getFloatValue(req.RecycledContentPercent, 0) * 0.25,
			"carbon_emissions":          (100 - getFloatValue(req.CarbonEmissions, 100)) * 0.2,
			"supply_chain_transparency": getFloatValue(req.SupplyChainTransparency, 0) * 0.4,
			"labor_standards":           getFloatValue(req.LaborStandardsCompliance, 0) * 0.3,
			"community_investment":      getFloatValue(req.CommunityInvestment, 0) * 0.2,
			"safety_performance":        (100 - getFloatValue(req.SafetyIncidentRate, 0)) * 0.1,
			"ethics_training":           getFloatValue(req.EthicsTrainingHours, 0) * 0.3,
			"board_diversity":           getFloatValue(req.BoardDiversity, 0) * 0.3,
			"data_privacy":              getFloatValue(req.DataPrivacyCompliance, 0) * 0.4,
		},
		"recommendations":     recommendations,
		"certification_level": ec.getCertificationLevel(overallScore),
		"assessment_method":   "AI-Generated",
		"generated_at":        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetESGRanking returns ESG ranking of passports
func (ec *ESGController) GetESGRanking(w http.ResponseWriter, r *http.Request) {
	// Extract user info from token
	claims, err := ec.extractUserClaims(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	manufacturer := r.URL.Query().Get("manufacturer")
	minScore, _ := strconv.ParseFloat(r.URL.Query().Get("min_score"), 64)

	// Get ranking from database
	ranking, err := ec.getESGRanking(limit, manufacturer, minScore)
	if err != nil {
		http.Error(w, "Failed to retrieve ESG ranking", http.StatusInternalServerError)
		return
	}

	// Log audit event
	ec.logAuditEvent(claims.UserID, claims.Role, "VIEW", "esg_ranking", "", nil, nil, r)

	response := map[string]interface{}{
		"ranking":     ranking,
		"total_count": len(ranking),
		"limit":       limit,
		"filters": map[string]interface{}{
			"manufacturer": manufacturer,
			"min_score":    minScore,
		},
		"generated_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ESG calculation methods
func (ec *ESGController) calculateEnvironmentalScore(req *ESGAssessmentRequest) float64 {
	scores := []float64{}

	if req.EnergyEfficiencyScore != nil {
		scores = append(scores, *req.EnergyEfficiencyScore)
	}
	if req.WaterUsageScore != nil {
		scores = append(scores, *req.WaterUsageScore)
	}
	if req.WasteManagementScore != nil {
		scores = append(scores, *req.WasteManagementScore)
	}
	if req.RenewableEnergyPercent != nil {
		scores = append(scores, *req.RenewableEnergyPercent)
	}

	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

func (ec *ESGController) calculateSocialScore(req *ESGAssessmentRequest) float64 {
	scores := []float64{}

	if req.LaborPracticesScore != nil {
		scores = append(scores, *req.LaborPracticesScore)
	}
	if req.CommunityImpactScore != nil {
		scores = append(scores, *req.CommunityImpactScore)
	}
	if req.HealthSafetyScore != nil {
		scores = append(scores, *req.HealthSafetyScore)
	}
	if req.HumanRightsScore != nil {
		scores = append(scores, *req.HumanRightsScore)
	}

	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

func (ec *ESGController) calculateGovernanceScore(req *ESGAssessmentRequest) float64 {
	scores := []float64{}

	if req.TransparencyScore != nil {
		scores = append(scores, *req.TransparencyScore)
	}
	if req.EthicsScore != nil {
		scores = append(scores, *req.EthicsScore)
	}
	if req.ComplianceScore != nil {
		scores = append(scores, *req.ComplianceScore)
	}
	if req.StakeholderEngagementScore != nil {
		scores = append(scores, *req.StakeholderEngagementScore)
	}

	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

// AI-based scoring methods (simplified algorithms)
func (ec *ESGController) calculateAIEnvironmentalScore(req *ESGScoreRequest) float64 {
	score := 0.0

	// Energy efficiency (30% weight)
	if req.RenewableEnergyPercent != nil {
		score += (*req.RenewableEnergyPercent / 100) * 30
	}

	// Waste management (25% weight)
	if req.WasteGenerated != nil {
		// Assume lower waste is better (inverted score)
		wasteScore := math.Max(0, 100-*req.WasteGenerated)
		score += (wasteScore / 100) * 25
	}

	// Recycled content (25% weight)
	if req.RecycledContentPercent != nil {
		score += (*req.RecycledContentPercent / 100) * 25
	}

	// Carbon emissions (20% weight)
	if req.CarbonEmissions != nil {
		// Assume lower emissions is better (inverted score)
		emissionScore := math.Max(0, 100-*req.CarbonEmissions)
		score += (emissionScore / 100) * 20
	}

	return math.Min(100, score)
}

func (ec *ESGController) calculateAISocialScore(req *ESGScoreRequest) float64 {
	score := 0.0

	// Supply chain transparency (40% weight)
	if req.SupplyChainTransparency != nil {
		score += (*req.SupplyChainTransparency / 100) * 40
	}

	// Labor standards (30% weight)
	if req.LaborStandardsCompliance != nil {
		score += (*req.LaborStandardsCompliance / 100) * 30
	}

	// Community investment (20% weight)
	if req.CommunityInvestment != nil {
		score += (*req.CommunityInvestment / 100) * 20
	}

	// Safety performance (10% weight)
	if req.SafetyIncidentRate != nil {
		// Assume lower incident rate is better
		safetyScore := math.Max(0, 100-*req.SafetyIncidentRate)
		score += (safetyScore / 100) * 10
	}

	return math.Min(100, score)
}

func (ec *ESGController) calculateAIGovernanceScore(req *ESGScoreRequest) float64 {
	score := 0.0

	// Ethics training (30% weight)
	if req.EthicsTrainingHours != nil {
		// Normalize training hours (assume 40 hours is max)
		trainingScore := math.Min(100, (*req.EthicsTrainingHours/40)*100)
		score += (trainingScore / 100) * 30
	}

	// Board diversity (30% weight)
	if req.BoardDiversity != nil {
		score += (*req.BoardDiversity / 100) * 30
	}

	// Data privacy compliance (40% weight)
	if req.DataPrivacyCompliance != nil {
		score += (*req.DataPrivacyCompliance / 100) * 40
	}

	return math.Min(100, score)
}

// Helper methods
func (ec *ESGController) calculateEnvironmentalScoreFromMetrics(metrics *db.ESGMetrics) float64 {
	scores := []float64{}

	if metrics.EnergyEfficiencyScore != nil {
		scores = append(scores, *metrics.EnergyEfficiencyScore)
	}
	if metrics.WaterUsageScore != nil {
		scores = append(scores, *metrics.WaterUsageScore)
	}
	if metrics.WasteManagementScore != nil {
		scores = append(scores, *metrics.WasteManagementScore)
	}
	if metrics.RenewableEnergyPercent != nil {
		scores = append(scores, *metrics.RenewableEnergyPercent)
	}

	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

func (ec *ESGController) calculateSocialScoreFromMetrics(metrics *db.ESGMetrics) float64 {
	scores := []float64{}

	if metrics.LaborPracticesScore != nil {
		scores = append(scores, *metrics.LaborPracticesScore)
	}
	if metrics.CommunityImpactScore != nil {
		scores = append(scores, *metrics.CommunityImpactScore)
	}
	if metrics.HealthSafetyScore != nil {
		scores = append(scores, *metrics.HealthSafetyScore)
	}
	if metrics.HumanRightsScore != nil {
		scores = append(scores, *metrics.HumanRightsScore)
	}

	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

func (ec *ESGController) calculateGovernanceScoreFromMetrics(metrics *db.ESGMetrics) float64 {
	scores := []float64{}

	if metrics.TransparencyScore != nil {
		scores = append(scores, *metrics.TransparencyScore)
	}
	if metrics.EthicsScore != nil {
		scores = append(scores, *metrics.EthicsScore)
	}
	if metrics.ComplianceScore != nil {
		scores = append(scores, *metrics.ComplianceScore)
	}
	if metrics.StakeholderEngagementScore != nil {
		scores = append(scores, *metrics.StakeholderEngagementScore)
	}

	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

func (ec *ESGController) getScoreBreakdown(metrics *db.ESGMetrics) map[string]float64 {
	breakdown := make(map[string]float64)

	if metrics.EnergyEfficiencyScore != nil {
		breakdown["energy_efficiency"] = *metrics.EnergyEfficiencyScore
	}
	if metrics.WaterUsageScore != nil {
		breakdown["water_usage"] = *metrics.WaterUsageScore
	}
	if metrics.WasteManagementScore != nil {
		breakdown["waste_management"] = *metrics.WasteManagementScore
	}
	if metrics.LaborPracticesScore != nil {
		breakdown["labor_practices"] = *metrics.LaborPracticesScore
	}
	if metrics.CommunityImpactScore != nil {
		breakdown["community_impact"] = *metrics.CommunityImpactScore
	}
	if metrics.TransparencyScore != nil {
		breakdown["transparency"] = *metrics.TransparencyScore
	}
	if metrics.ComplianceScore != nil {
		breakdown["compliance"] = *metrics.ComplianceScore
	}

	return breakdown
}

func (ec *ESGController) generateRecommendations(metrics *db.ESGMetrics) []string {
	recommendations := []string{}

	if metrics.EnergyEfficiencyScore != nil && *metrics.EnergyEfficiencyScore < 70 {
		recommendations = append(recommendations, "Consider investing in renewable energy sources to improve energy efficiency")
	}

	if metrics.WasteManagementScore != nil && *metrics.WasteManagementScore < 70 {
		recommendations = append(recommendations, "Implement circular economy practices to reduce waste generation")
	}

	if metrics.LaborPracticesScore != nil && *metrics.LaborPracticesScore < 70 {
		recommendations = append(recommendations, "Enhance worker safety programs and fair labor practices")
	}

	if metrics.TransparencyScore != nil && *metrics.TransparencyScore < 70 {
		recommendations = append(recommendations, "Increase supply chain transparency and stakeholder communication")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Maintain current excellent ESG practices")
	}

	return recommendations
}

func (ec *ESGController) generateAIRecommendations(req *ESGScoreRequest, envScore, socialScore, govScore float64) []string {
	recommendations := []string{}

	if envScore < 70 {
		if req.RenewableEnergyPercent != nil && *req.RenewableEnergyPercent < 50 {
			recommendations = append(recommendations, "Increase renewable energy usage to improve environmental score")
		}
		if req.RecycledContentPercent != nil && *req.RecycledContentPercent < 30 {
			recommendations = append(recommendations, "Increase recycled content in products")
		}
	}

	if socialScore < 70 {
		recommendations = append(recommendations, "Improve supply chain transparency and labor practices")
	}

	if govScore < 70 {
		recommendations = append(recommendations, "Enhance governance practices and compliance measures")
	}

	return recommendations
}

func (ec *ESGController) getCertificationLevel(score float64) string {
	if score >= 90 {
		return "Platinum"
	} else if score >= 80 {
		return "Gold"
	} else if score >= 70 {
		return "Silver"
	} else if score >= 60 {
		return "Bronze"
	}
	return "Not Certified"
}

// Database helper methods
func (ec *ESGController) passportExists(passportID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM aluminium_passports WHERE passport_id = $1)`
	var exists bool
	err := db.DB.QueryRow(query, passportID).Scan(&exists)
	return exists, err
}

func (ec *ESGController) createESGMetrics(metrics *db.ESGMetrics) (int, error) {
	query := `
		INSERT INTO esg_metrics (
			passport_id, carbon_footprint, energy_efficiency_score, water_usage_score, waste_management_score,
			renewable_energy_percent, labor_practices_score, community_impact_score, health_safety_score,
			human_rights_score, transparency_score, ethics_score, compliance_score, stakeholder_engagement_score,
			overall_esg_score, assessment_date, assessor_id, assessment_methodology, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		) RETURNING id`

	var esgID int
	err := db.DB.QueryRow(
		query,
		metrics.PassportID, metrics.CarbonFootprint, metrics.EnergyEfficiencyScore, metrics.WaterUsageScore,
		metrics.WasteManagementScore, metrics.RenewableEnergyPercent, metrics.LaborPracticesScore,
		metrics.CommunityImpactScore, metrics.HealthSafetyScore, metrics.HumanRightsScore,
		metrics.TransparencyScore, metrics.EthicsScore, metrics.ComplianceScore,
		metrics.StakeholderEngagementScore, metrics.OverallESGScore, metrics.AssessmentDate,
		metrics.AssessorID, metrics.AssessmentMethodology, metrics.Notes, metrics.CreatedAt, metrics.UpdatedAt,
	).Scan(&esgID)

	return esgID, err
}

func (ec *ESGController) getESGMetricsByPassportID(passportID string) (*db.ESGMetrics, error) {
	query := `
		SELECT id, passport_id, carbon_footprint, energy_efficiency_score, water_usage_score, waste_management_score,
		       renewable_energy_percent, labor_practices_score, community_impact_score, health_safety_score,
		       human_rights_score, transparency_score, ethics_score, compliance_score, stakeholder_engagement_score,
		       overall_esg_score, assessment_date, assessor_id, assessment_methodology, notes, created_at, updated_at
		FROM esg_metrics 
		WHERE passport_id = $1
		ORDER BY created_at DESC 
		LIMIT 1`

	metrics := &db.ESGMetrics{}
	err := db.DB.QueryRow(query, passportID).Scan(
		&metrics.ID, &metrics.PassportID, &metrics.CarbonFootprint, &metrics.EnergyEfficiencyScore,
		&metrics.WaterUsageScore, &metrics.WasteManagementScore, &metrics.RenewableEnergyPercent,
		&metrics.LaborPracticesScore, &metrics.CommunityImpactScore, &metrics.HealthSafetyScore,
		&metrics.HumanRightsScore, &metrics.TransparencyScore, &metrics.EthicsScore,
		&metrics.ComplianceScore, &metrics.StakeholderEngagementScore, &metrics.OverallESGScore,
		&metrics.AssessmentDate, &metrics.AssessorID, &metrics.AssessmentMethodology,
		&metrics.Notes, &metrics.CreatedAt, &metrics.UpdatedAt,
	)

	return metrics, err
}

func (ec *ESGController) updatePassportESGScores(passportID string, envScore, socialScore, govScore, overallScore float64, userID int) error {
	query := `
		UPDATE aluminium_passports 
		SET environmental_score = $1, social_score = $2, governance_score = $3, esg_score = $4, 
		    esg_last_updated = $5, updated_by = $6, updated_at = $7
		WHERE passport_id = $8`

	_, err := db.DB.Exec(query, envScore, socialScore, govScore, overallScore, time.Now(), userID, time.Now(), passportID)
	return err
}

func (ec *ESGController) getESGRanking(limit int, manufacturer string, minScore float64) ([]map[string]interface{}, error) {
	whereClauses := []string{"e.overall_esg_score IS NOT NULL"}
	args := []interface{}{}
	argIndex := 1

	if manufacturer != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.manufacturer ILIKE $%d", argIndex))
		args = append(args, "%"+manufacturer+"%")
		argIndex++
	}

	if minScore > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("e.overall_esg_score >= $%d", argIndex))
		args = append(args, minScore)
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT p.passport_id, p.manufacturer, p.origin, e.overall_esg_score, 
		       e.assessment_date, p.created_at,
		       ROW_NUMBER() OVER (ORDER BY e.overall_esg_score DESC) as rank
		FROM aluminium_passports p
		JOIN esg_metrics e ON p.passport_id = e.passport_id
		WHERE %s
		ORDER BY e.overall_esg_score DESC
		LIMIT $%d`, whereClause, argIndex)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranking []map[string]interface{}
	for rows.Next() {
		var passportID, manufacturer, origin string
		var esgScore float64
		var assessmentDate, createdAt time.Time
		var rank int

		err := rows.Scan(&passportID, &manufacturer, &origin, &esgScore, &assessmentDate, &createdAt, &rank)
		if err != nil {
			return nil, err
		}

		ranking = append(ranking, map[string]interface{}{
			"rank":                rank,
			"passport_id":         passportID,
			"manufacturer":        manufacturer,
			"origin":              origin,
			"esg_score":           esgScore,
			"certification_level": ec.getCertificationLevel(esgScore),
			"assessment_date":     assessmentDate,
			"created_at":          createdAt,
		})
	}

	return ranking, nil
}

// Helper methods
func (ec *ESGController) extractUserClaims(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return auth.ValidateToken(tokenString)
}

func (ec *ESGController) hasRole(userRole string, allowedRoles []string) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func (ec *ESGController) logAuditEvent(userID int, userRole, action, resourceType, resourceID string, oldValues, newValues interface{}, r *http.Request) {
	// Implementation would log to audit_logs table
}

func timePtr(t time.Time) *time.Time {
	return &t
}
