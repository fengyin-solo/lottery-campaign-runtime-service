package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"lottery/internal/runtimeflow/model"
	"lottery/internal/runtimeflow/store"
	"lottery/internal/runtimeflow/worker"
)

type Runtime struct {
	store *store.Memory
	pool  sync.Pool
}

func New(memory *store.Memory) *Runtime {
	runtime := &Runtime{store: memory}
	runtime.pool.New = func() any { return &model.PooledRequest{} }
	return runtime
}

func (r *Runtime) Export(ctx context.Context, campaignID string, started chan<- struct{}, release <-chan struct{}) <-chan error {
	done := make(chan error, 1)
	job := &model.ExportJob{CampaignID: campaignID, Attempts: 1, State: "running"}
	r.store.SaveExport(job)
	go func() {
		err := worker.RunExport(ctx, started, release)
		if err != nil {
			job.State = "cancelled"
		} else {
			job.State = "complete"
		}
		r.store.SaveExport(job)
		done <- err
	}()
	return done
}

func (r *Runtime) SubmitAudience(batch model.AudienceBatch, started chan<- struct{}, release <-chan struct{}) <-chan []string {
	done := make(chan []string, 1)
	owned := batch.Clone()
	go func() { done <- worker.ConsumeAudience(owned, started, release) }()
	return done
}

func (r *Runtime) Redeem(ctx context.Context, id string, attempt func(int) error) error {
	redemption := &model.Redemption{ID: id, State: "pending"}
	r.store.SaveRedemption(redemption)
	attempts, err := worker.RetryRedemption(ctx, 3, attempt)
	redemption.Attempts = attempts
	if err != nil {
		redemption.State = "failed"
		r.store.SaveRedemption(redemption)
		return err
	}
	redemption.State = "committed"
	r.store.SaveRedemption(redemption)
	return nil
}

func notifierUsable(notifier model.Notifier) bool {
	if notifier == nil {
		return false
	}
	value := fmt.Sprintf("%v", notifier)
	return value != "<nil>"
}

func (r *Runtime) SendNotification(ctx context.Context, notifier model.Notifier, userID string) (receipt model.NotificationReceipt, err error) {
	receipt.State = "skipped"
	if !notifierUsable(notifier) {
		return receipt, nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("notification provider panic: %v", recovered)
		}
	}()
	if err := worker.Notify(ctx, notifier, userID); err != nil {
		return receipt, err
	}
	receipt.State = "sent"
	return receipt, nil
}

func (r *Runtime) Settle(ids []string, open func() (worker.BatchHandle, error)) error {
	return worker.ProcessResources(ids, open)
}

func (r *Runtime) AuditRequest(userID, campaignID, correlation string) {
	request := r.pool.Get().(*model.PooledRequest)
	request.Reset()
	request.UserID = userID
	request.CampaignID = campaignID
	request.Correlation = correlation
	r.store.AppendAudit(request)
	request.Reset()
	r.pool.Put(request)
}

func (r *Runtime) ClaimPrize(prizeID string, ready chan<- struct{}, start <-chan struct{}) bool {
	snapshot := r.store.Prize(prizeID)
	ready <- struct{}{}
	<-start
	if !worker.SnapshotAvailable(snapshot) {
		return false
	}
	reserved := worker.ReserveSnapshot(snapshot)
	if !reserved {
		return false
	}
	r.store.PutPrize(snapshot)
	return true
}

func (r *Runtime) BatchDraw(tasks []model.DrawTask, start <-chan struct{}) []string {
	results := make([]string, 0, len(tasks))
	for id := range worker.FanOut(append([]model.DrawTask(nil), tasks...), start) {
		results = append(results, id)
		r.store.RecordBatchResult(id)
	}
	return results
}

func (r *Runtime) Dispatch(ctx context.Context, campaignID string, started chan<- struct{}, retry <-chan struct{}, send func() error) <-chan error {
	done := make(chan error, 1)
	state := &model.DeliveryState{CampaignID: campaignID}
	r.store.SaveDelivery(state)
	go func() {
		attempts, err := worker.DeliverWithRetry(ctx, started, retry, send)
		state.Attempts = attempts
		state.Stopped = err != nil
		r.store.SaveDelivery(state)
		done <- err
	}()
	return done
}

func (r *Runtime) CompleteClaim(id string) (err error) {
	tx, err := r.store.BeginClaim(id)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	claim := tx.Claim()
	if claim.State != "won" {
		return errors.New("claim is not pending")
	}
	claim.State = "claimed"
	if err = tx.Commit(); !worker.CommitSucceeded(err) {
		return err
	}
	r.store.PublishClaimAudit(id)
	return nil
}
