package store

import (
	"errors"
	"sync"

	"lottery/internal/runtimeflow/model"
)

type Memory struct {
	mu          sync.Mutex
	exports     map[string]*model.ExportJob
	redemptions map[string]*model.Redemption
	audits      []model.PooledRequest
	prizes      map[string]*model.PrizeSnapshot
	deliveries  map[string]*model.DeliveryState
	claims      map[string]*model.Claim
	claimAudits []string
	batchRuns   map[string]int
	commitError error
}

func NewMemory() *Memory {
	return &Memory{
		exports: make(map[string]*model.ExportJob), redemptions: make(map[string]*model.Redemption),
		prizes: make(map[string]*model.PrizeSnapshot), deliveries: make(map[string]*model.DeliveryState),
		claims:    make(map[string]*model.Claim),
		batchRuns: make(map[string]int),
	}
}

func (m *Memory) RecordBatchResult(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batchRuns[id]++
}

func (m *Memory) BatchResultCount(id string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.batchRuns[id]
}

func (m *Memory) SaveExport(job *model.ExportJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exports[job.CampaignID] = job.Clone()
}

func (m *Memory) Export(id string) *model.ExportJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exports[id].Clone()
}

func (m *Memory) SaveRedemption(redemption *model.Redemption) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.redemptions[redemption.ID] = redemption.Clone()
}

func (m *Memory) Redemption(id string) *model.Redemption {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.redemptions[id].Clone()
}

func (m *Memory) AppendAudit(request *model.PooledRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.audits = append(m.audits, request.Clone())
}

func (m *Memory) Audits() []model.PooledRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]model.PooledRequest(nil), m.audits...)
}

func (m *Memory) PutPrize(prize *model.PrizeSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prizes[prize.PrizeID] = prize.Clone()
}

func (m *Memory) Prize(id string) *model.PrizeSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.prizes[id].Clone()
}

func (m *Memory) ReservePrize(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	prize := m.prizes[id]
	if prize == nil || prize.Remaining == 0 {
		return false
	}
	prize.Remaining--
	return true
}

func (m *Memory) SaveDelivery(state *model.DeliveryState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deliveries[state.CampaignID] = state
}

func (m *Memory) Delivery(id string) *model.DeliveryState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deliveries[id]
}

func (m *Memory) PutClaim(claim *model.Claim) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims[claim.ID] = claim.Clone()
}

func (m *Memory) Claim(id string) *model.Claim {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.claims[id].Clone()
}

func (m *Memory) SetCommitError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commitError = err
}

func (m *Memory) CommitClaim(claim *model.Claim) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.commitError != nil {
		err := m.commitError
		m.commitError = nil
		return err
	}
	m.claims[claim.ID] = claim.Clone()
	return nil
}

func (m *Memory) PublishClaimAudit(claimID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claimAudits = append(m.claimAudits, claimID)
}

func (m *Memory) ClaimAudits() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.claimAudits...)
}

func (m *Memory) BeginClaim(id string) (*ClaimTx, error) {
	claim := m.Claim(id)
	if claim == nil {
		return nil, errors.New("claim not found")
	}
	return &ClaimTx{memory: m, original: claim.Clone(), current: claim.Clone()}, nil
}

type ClaimTx struct {
	memory   *Memory
	original *model.Claim
	current  *model.Claim
	done     bool
}

func (tx *ClaimTx) Claim() *model.Claim { return tx.current }

func (tx *ClaimTx) Commit() error {
	if tx.done {
		return errors.New("transaction closed")
	}
	if err := tx.memory.CommitClaim(tx.current); err != nil {
		return err
	}
	tx.done = true
	return nil
}

func (tx *ClaimTx) Rollback() error {
	if tx.done {
		return nil
	}
	tx.done = true
	return nil
}
