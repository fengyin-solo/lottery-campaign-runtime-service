package model

import (
	"context"
	"errors"
	"sync/atomic"
)

type ExportJob struct {
	CampaignID string
	Attempts   int
	State      string
}

func (j *ExportJob) Clone() *ExportJob {
	if j == nil {
		return nil
	}
	copy := *j
	return &copy
}

type AudienceBatch struct {
	CampaignID string
	UserIDs    []string
}

func NewAudienceBatch(campaignID string, userIDs []string) AudienceBatch {
	return AudienceBatch{CampaignID: campaignID, UserIDs: append([]string(nil), userIDs...)}
}

func (b AudienceBatch) Clone() AudienceBatch {
	b.UserIDs = append([]string(nil), b.UserIDs...)
	return b
}

var ErrTemporary = errors.New("temporary redemption failure")

type Redemption struct {
	ID       string
	Attempts int
	State    string
}

func (r *Redemption) Clone() *Redemption {
	if r == nil {
		return nil
	}
	copy := *r
	return &copy
}

type Notifier interface {
	Notify(context.Context, string) error
}

type NotificationReceipt struct {
	State string
}

type ResourceTracker struct {
	open    atomic.Int32
	maximum atomic.Int32
}

func (t *ResourceTracker) Opened() {
	now := t.open.Add(1)
	for {
		old := t.maximum.Load()
		if now <= old || t.maximum.CompareAndSwap(old, now) {
			return
		}
	}
}

func (t *ResourceTracker) Closed()        {}
func (t *ResourceTracker) Open() int32    { return t.open.Load() }
func (t *ResourceTracker) Maximum() int32 { return t.maximum.Load() }

type PooledRequest struct {
	UserID      string
	CampaignID  string
	Correlation string
}

func (r *PooledRequest) Reset() {
	r.UserID = ""
	r.CampaignID = ""
	r.Correlation = ""
}

func (r *PooledRequest) Clone() PooledRequest {
	if r == nil {
		return PooledRequest{}
	}
	return *r
}

type PrizeSnapshot struct {
	PrizeID   string
	Remaining int
}

func (p *PrizeSnapshot) Clone() *PrizeSnapshot {
	if p == nil {
		return nil
	}
	copy := *p
	return &copy
}

type DrawTask struct {
	ID     string
	Labels []string
}

func (t DrawTask) Clone() DrawTask {
	t.Labels = append([]string(nil), t.Labels...)
	return t
}

type DeliveryState struct {
	CampaignID string
	Attempts   int
	Stopped    bool
}

func (s *DeliveryState) Clone() *DeliveryState {
	if s == nil {
		return nil
	}
	copy := *s
	return &copy
}

type Claim struct {
	ID      string
	PrizeID string
	State   string
}

func (c *Claim) Clone() *Claim {
	if c == nil {
		return nil
	}
	copy := *c
	return &copy
}
