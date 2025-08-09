package services

import (
	"errors"
	"sort"
	"sync"
)

// Demo domain models mirror the AluminiumPassportDemo.sol structures

type DemoOnboardingRequest struct {
	OrgID          string   `json:"orgId"`
	Wallet         string   `json:"wallet"`
	KYCCID         string   `json:"kycCid"`
	MetaCID        string   `json:"metaCid"`
	RolesRequested []string `json:"rolesRequested"`
	Exists         bool     `json:"exists"`
	Approved       bool     `json:"approved"`
	RequestedAt    int64    `json:"requestedAt"`
	ApprovedAt     int64    `json:"approvedAt"`
}

type DemoUpstreamBatch struct {
	BatchID      string `json:"batchId"`
	CID          string `json:"cid"`
	RegisteredBy string `json:"registeredBy"`
	Timestamp    int64  `json:"timestamp"`
	Exists       bool   `json:"exists"`
}

type DemoPassport struct {
	ID              uint64 `json:"id"`
	OrgID           string `json:"orgId"`
	UpstreamBatchID string `json:"upstreamBatchId"`
	MetaCID         string `json:"metaCid"`
	ParentID        uint64 `json:"parentId"`
	CreatedBy       string `json:"createdBy"`
	CreatedAt       int64  `json:"createdAt"`
	Exists          bool   `json:"exists"`
}

type DemoPlacedOnMarket struct {
	CountryCode string `json:"countryCode"`
	DateISO     string `json:"dateISO"`
	CID         string `json:"cid"`
	RecordedBy  string `json:"recordedBy"`
	Timestamp   int64  `json:"timestamp"`
	Exists      bool   `json:"exists"`
}

type DemoAttestation struct {
	CID        string `json:"cid"`
	AttestedBy string `json:"attestedBy"`
	Timestamp  int64  `json:"timestamp"`
}

type DemoRecovery struct {
	RecoveryPercent uint8  `json:"recoveryPercent"`
	Quality         string `json:"quality"`
	CID             string `json:"cid"`
	RecordedBy      string `json:"recordedBy"`
	Timestamp       int64  `json:"timestamp"`
}

type DemoStore struct {
	mu              sync.Mutex
	onboardingByOrg map[string]*DemoOnboardingRequest
	walletToOrg     map[string]string
	orgSuspended    map[string]bool
	upstreamByID    map[string]*DemoUpstreamBatch
	passports       map[uint64]*DemoPassport
	placed          map[uint64]*DemoPlacedOnMarket
	attestations    map[uint64][]*DemoAttestation
	recovery        map[uint64][]*DemoRecovery
	nextPassportID  uint64
}

func NewDemoStore() *DemoStore {
	return &DemoStore{
		onboardingByOrg: make(map[string]*DemoOnboardingRequest),
		walletToOrg:     make(map[string]string),
		orgSuspended:    make(map[string]bool),
		upstreamByID:    make(map[string]*DemoUpstreamBatch),
		passports:       make(map[uint64]*DemoPassport),
		placed:          make(map[uint64]*DemoPlacedOnMarket),
		attestations:    make(map[uint64][]*DemoAttestation),
		recovery:        make(map[uint64][]*DemoRecovery),
		nextPassportID:  1,
	}
}

func (s *DemoStore) RequestOnboarding(now int64, orgID, wallet, kycCID, metaCID string, roles []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onboardingByOrg[orgID] = &DemoOnboardingRequest{
		OrgID:          orgID,
		Wallet:         wallet,
		KYCCID:         kycCID,
		MetaCID:        metaCID,
		RolesRequested: roles,
		Exists:         true,
		Approved:       false,
		RequestedAt:    now,
	}
}

func (s *DemoStore) ApproveOnboarding(now int64, orgID string, roles []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.onboardingByOrg[orgID]
	if !ok || !r.Exists {
		return errors.New("no request")
	}
	if r.Approved {
		return errors.New("already approved")
	}
	r.Approved = true
	r.ApprovedAt = now
	s.walletToOrg[r.Wallet] = orgID
	return nil
}

func (s *DemoStore) RegisterUpstream(now int64, batchID, cid, registeredBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.upstreamByID[batchID]; exists {
		return errors.New("batch exists")
	}
	s.upstreamByID[batchID] = &DemoUpstreamBatch{
		BatchID:      batchID,
		CID:          cid,
		RegisteredBy: registeredBy,
		Timestamp:    now,
		Exists:       true,
	}
	return nil
}

func (s *DemoStore) CreatePassport(now int64, orgID, upstreamBatchID, metaCID, createdBy string) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if upstreamBatchID != "" {
		if _, ok := s.upstreamByID[upstreamBatchID]; !ok {
			return 0, errors.New("no upstream")
		}
	}
	id := s.nextPassportID
	s.nextPassportID++
	s.passports[id] = &DemoPassport{
		ID:              id,
		OrgID:           orgID,
		UpstreamBatchID: upstreamBatchID,
		MetaCID:         metaCID,
		ParentID:        0,
		CreatedBy:       createdBy,
		CreatedAt:       now,
		Exists:          true,
	}
	return id, nil
}

func (s *DemoStore) RecordPlaced(now int64, passportID uint64, countryCode, dateISO, cid, recordedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.passports[passportID]; !ok {
		return errors.New("no passport")
	}
	s.placed[passportID] = &DemoPlacedOnMarket{
		CountryCode: countryCode,
		DateISO:     dateISO,
		CID:         cid,
		RecordedBy:  recordedBy,
		Timestamp:   now,
		Exists:      true,
	}
	return nil
}

func (s *DemoStore) AddAttestation(now int64, passportID uint64, cid, attestedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.passports[passportID]; !ok {
		return errors.New("no passport")
	}
	s.attestations[passportID] = append(s.attestations[passportID], &DemoAttestation{CID: cid, AttestedBy: attestedBy, Timestamp: now})
	return nil
}

func (s *DemoStore) RecordRecovery(now int64, passportID uint64, pct uint8, quality, cid, recordedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.passports[passportID]; !ok {
		return errors.New("no passport")
	}
	s.recovery[passportID] = append(s.recovery[passportID], &DemoRecovery{RecoveryPercent: pct, Quality: quality, CID: cid, RecordedBy: recordedBy, Timestamp: now})
	return nil
}

func (s *DemoStore) SpawnSecondary(now int64, parentID uint64, metaCID, createdBy string) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.passports[parentID]
	if !ok || !p.Exists {
		return 0, errors.New("no parent")
	}
	id := s.nextPassportID
	s.nextPassportID++
	s.passports[id] = &DemoPassport{
		ID:              id,
		OrgID:           p.OrgID,
		UpstreamBatchID: p.UpstreamBatchID,
		MetaCID:         metaCID,
		ParentID:        parentID,
		CreatedBy:       createdBy,
		CreatedAt:       now,
		Exists:          true,
	}
	return id, nil
}

// Public view

type DemoPublicView struct {
	OrgID           string `json:"orgId"`
	UpstreamBatchID string `json:"upstreamBatchId"`
	PassportMetaCID string `json:"passportMetaCid"`
	Placed          bool   `json:"placed"`
	CountryCode     string `json:"countryCode"`
	DateISO         string `json:"dateISO"`
	PlacedCID       string `json:"placedCid"`
	HasAttestation  bool   `json:"hasAttestation"`
}

func (s *DemoStore) GetPublicView(passportID uint64) (*DemoPublicView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.passports[passportID]
	if !ok || !p.Exists {
		return nil, errors.New("no passport")
	}
	m, mok := s.placed[passportID]
	att := s.attestations[passportID]
	pv := &DemoPublicView{
		OrgID:           p.OrgID,
		UpstreamBatchID: p.UpstreamBatchID,
		PassportMetaCID: p.MetaCID,
		Placed:          mok && m.Exists,
		HasAttestation:  len(att) > 0,
	}
	if mok && m.Exists {
		pv.CountryCode = m.CountryCode
		pv.DateISO = m.DateISO
		pv.PlacedCID = m.CID
	}
	return pv, nil
}

// List passports (for importer incoming/inventory pages)
func (s *DemoStore) ListPassports() []*DemoPassport {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]uint64, 0, len(s.passports))
	for id := range s.passports {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make([]*DemoPassport, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.passports[id])
	}
	return out
}
