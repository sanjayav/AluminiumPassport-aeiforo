package demo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"aluminium-passport/internal/services"

	"github.com/gorilla/mux"
)

type Handler struct {
	store *services.DemoStore
}

func NewRouter() *mux.Router {
	h := &Handler{store: services.NewDemoStore()}
	r := mux.NewRouter()

	// Health
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"demo"}`))
	}).Methods("GET")

	api := r.PathPrefix("/api/demo").Subrouter()
	api.HandleFunc("/onboard", h.requestOnboarding).Methods("POST")
	api.HandleFunc("/approve", h.approveOnboarding).Methods("POST")
	api.HandleFunc("/upstream", h.registerUpstream).Methods("POST")
	api.HandleFunc("/passports", h.createPassport).Methods("POST")
	api.HandleFunc("/passports", h.listPassports).Methods("GET")
	api.HandleFunc("/market", h.recordPlacement).Methods("POST")
	api.HandleFunc("/attest", h.addAttestation).Methods("POST")
	api.HandleFunc("/recycle", h.recordRecovery).Methods("POST")
	api.HandleFunc("/secondary", h.spawnSecondary).Methods("POST")
	api.HandleFunc("/public/{id}", h.getPublicView).Methods("GET")

	return r
}

func (h *Handler) requestOnboarding(w http.ResponseWriter, r *http.Request) {
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
	h.store.RequestOnboarding(time.Now().Unix(), req.OrgID, req.Wallet, req.KYCCID, req.MetaCID, req.Roles)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "pending", "orgId": req.OrgID})
}

func (h *Handler) approveOnboarding(w http.ResponseWriter, r *http.Request) {
	type reqT struct {
		OrgID string   `json:"orgId"`
		Roles []string `json:"roles"`
	}
	var req reqT
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.store.ApproveOnboarding(time.Now().Unix(), req.OrgID, req.Roles); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "approved", "orgId": req.OrgID})
}

func (h *Handler) registerUpstream(w http.ResponseWriter, r *http.Request) {
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
	if err := h.store.RegisterUpstream(time.Now().Unix(), req.BatchID, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"batchId": req.BatchID, "cid": req.CID})
}

func (h *Handler) createPassport(w http.ResponseWriter, r *http.Request) {
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
	id, err := h.store.CreatePassport(time.Now().Unix(), req.OrgID, req.UpstreamBatchID, req.MetaCID, req.By)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"passportId": id})
}

func (h *Handler) listPassports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.store.ListPassports())
}

func (h *Handler) recordPlacement(w http.ResponseWriter, r *http.Request) {
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
	if err := h.store.RecordPlaced(time.Now().Unix(), req.PassportID, req.Country, req.DateISO, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "placed"})
}

func (h *Handler) addAttestation(w http.ResponseWriter, r *http.Request) {
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
	if err := h.store.AddAttestation(time.Now().Unix(), req.PassportID, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "attested"})
}

func (h *Handler) recordRecovery(w http.ResponseWriter, r *http.Request) {
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
	if err := h.store.RecordRecovery(time.Now().Unix(), req.PassportID, req.Percent, req.Quality, req.CID, req.By); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

func (h *Handler) spawnSecondary(w http.ResponseWriter, r *http.Request) {
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
	id, err := h.store.SpawnSecondary(time.Now().Unix(), req.ParentID, req.MetaCID, req.By)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint64{"passportId": id})
}

func (h *Handler) getPublicView(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	pv, err := h.store.GetPublicView(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pv)
}
