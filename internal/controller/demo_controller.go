package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"aluminium-passport/internal/services"

	"github.com/gorilla/mux"
)

type DemoController struct {
	store *services.DemoStore
}

func NewDemoController() *DemoController {
	return &DemoController{store: services.NewDemoStore()}
}

// POST /api/demo/onboard
func (dc *DemoController) RequestOnboarding(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		OrgID   string   `json:"orgId"`
		Wallet  string   `json:"wallet"`
		KYCCID  string   `json:"kycCid"`
		MetaCID string   `json:"metaCid"`
		Roles   []string `json:"roles"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	dc.store.RequestOnboarding(time.Now().Unix(), req.OrgID, req.Wallet, req.KYCCID, req.MetaCID, req.Roles)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "pending", "orgId": req.OrgID})
}

// POST /api/demo/approve
func (dc *DemoController) ApproveOnboarding(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		OrgID string   `json:"orgId"`
		Roles []string `json:"roles"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := dc.store.ApproveOnboarding(time.Now().Unix(), req.OrgID, req.Roles); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "approved", "orgId": req.OrgID})
}

// POST /api/demo/upstream
func (dc *DemoController) RegisterUpstream(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		BatchID string `json:"batchId"`
		CID     string `json:"cid"`
		By      string `json:"by"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := dc.store.RegisterUpstream(time.Now().Unix(), req.BatchID, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"batchId": req.BatchID, "cid": req.CID})
}

// POST /api/demo/passports
func (dc *DemoController) CreatePassport(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		OrgID           string `json:"orgId"`
		UpstreamBatchID string `json:"upstreamBatchId"`
		MetaCID         string `json:"metaCid"`
		By              string `json:"by"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := dc.store.CreatePassport(time.Now().Unix(), req.OrgID, req.UpstreamBatchID, req.MetaCID, req.By)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"passportId": id})
}

// GET /api/demo/passports
func (dc *DemoController) ListPassports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dc.store.ListPassports())
}

// POST /api/demo/market
func (dc *DemoController) RecordPlacement(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		PassportID uint64 `json:"passportId"`
		Country    string `json:"country"`
		DateISO    string `json:"dateISO"`
		CID        string `json:"cid"`
		By         string `json:"by"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := dc.store.RecordPlaced(time.Now().Unix(), req.PassportID, req.Country, req.DateISO, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "placed"})
}

// POST /api/demo/attest
func (dc *DemoController) AddAttestation(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		PassportID uint64 `json:"passportId"`
		CID        string `json:"cid"`
		By         string `json:"by"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := dc.store.AddAttestation(time.Now().Unix(), req.PassportID, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "attested"})
}

// POST /api/demo/recycle
func (dc *DemoController) RecordRecovery(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		PassportID uint64 `json:"passportId"`
		Percent    uint8  `json:"percent"`
		Quality    string `json:"quality"`
		CID        string `json:"cid"`
		By         string `json:"by"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := dc.store.RecordRecovery(time.Now().Unix(), req.PassportID, req.Percent, req.Quality, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// POST /api/demo/secondary
func (dc *DemoController) SpawnSecondary(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		ParentID uint64 `json:"parentId"`
		MetaCID  string `json:"metaCid"`
		By       string `json:"by"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	id, err := dc.store.SpawnSecondary(time.Now().Unix(), req.ParentID, req.MetaCID, req.By)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint64{"passportId": id})
}

// GET /api/demo/public/{id}
func (dc *DemoController) GetPublicView(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	pv, err := dc.store.GetPublicView(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pv)
}
