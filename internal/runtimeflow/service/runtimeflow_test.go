package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"lottery/internal/runtimeflow/model"
	"lottery/internal/runtimeflow/store"
	"lottery/internal/runtimeflow/worker"
)

func TestCancelledExportStopsAndNextExportIsolated(t *testing.T) {
	memory := store.NewMemory()
	runtime := New(memory)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	release := make(chan struct{})
	done := runtime.Export(ctx, "campaign-old", started, release)
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled export returned %v", err)
	}
	if job := memory.Export("campaign-old"); job == nil || job.State != "cancelled" {
		t.Fatalf("cancelled export state = %#v", job)
	}
	started2 := make(chan struct{})
	release2 := make(chan struct{})
	done2 := runtime.Export(context.Background(), "campaign-new", started2, release2)
	<-started2
	close(release2)
	if err := <-done2; err != nil {
		t.Fatalf("next export inherited cancellation: %v", err)
	}
	if job := memory.Export("campaign-new"); job == nil || job.State != "complete" {
		t.Fatalf("next export state = %#v", job)
	}
}

func TestAudienceBatchKeepsSubmittedUsers(t *testing.T) {
	runtime := New(store.NewMemory())
	users := []string{"u-1", "u-2"}
	started := make(chan struct{})
	release := make(chan struct{})
	done := runtime.SubmitAudience(model.NewAudienceBatch("campaign-a", users), started, release)
	<-started
	users[0] = "u-reused"
	users = append(users[:0], "u-next")
	close(release)
	if got := <-done; !reflect.DeepEqual(got, []string{"u-1", "u-2"}) {
		t.Fatalf("submitted audience changed after caller reuse: %v", got)
	}
	started2 := make(chan struct{})
	release2 := make(chan struct{})
	next := runtime.SubmitAudience(model.NewAudienceBatch("campaign-b", []string{"u-3"}), started2, release2)
	<-started2
	close(release2)
	if got := <-next; !reflect.DeepEqual(got, []string{"u-3"}) {
		t.Fatalf("next campaign inherited prior audience: %v", got)
	}
}

func TestWrappedTemporaryRedemptionRetriesWithoutDuplicateCommit(t *testing.T) {
	memory := store.NewMemory()
	runtime := New(memory)
	attempts := 0
	err := runtime.Redeem(context.Background(), "redeem-1", func(attempt int) error {
		attempts++
		if attempt == 1 {
			return fmt.Errorf("gateway: %w", model.ErrTemporary)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("redeem returned %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempt count = %d", attempts)
	}
	if got := memory.Redemption("redeem-1"); got == nil || got.Attempts != 2 || got.State != "committed" {
		t.Fatalf("redemption = %#v", got)
	}
}

type nilNotifier struct{}

func (n *nilNotifier) Notify(context.Context, string) error {
	if n == nil {
		panic("nil notifier dereference")
	}
	return nil
}

func TestTypedNilNotifierSkipsAndLaterSendStillWorks(t *testing.T) {
	runtime := New(store.NewMemory())
	var provider *nilNotifier
	var notifier model.Notifier = provider
	receipt, err := runtime.SendNotification(context.Background(), notifier, "u-1")
	if err != nil || receipt.State != "skipped" {
		t.Fatalf("typed nil result = %#v, %v", receipt, err)
	}
	receipt, err = runtime.SendNotification(context.Background(), &nilNotifier{}, "u-2")
	if err != nil || receipt.State != "sent" {
		t.Fatalf("later send result = %#v, %v", receipt, err)
	}
}

type trackedHandle struct {
	tracker *model.ResourceTracker
}

func (h *trackedHandle) Process(string) error { return nil }
func (h *trackedHandle) Close() error {
	h.tracker.Closed()
	return nil
}

func TestLargeSettlementClosesEachResourcePromptly(t *testing.T) {
	runtime := New(store.NewMemory())
	tracker := &model.ResourceTracker{}
	ids := make([]string, 128)
	for i := range ids {
		ids[i] = fmt.Sprintf("winner-%d", i)
	}
	err := runtime.Settle(ids, func() (worker.BatchHandle, error) {
		tracker.Opened()
		return &trackedHandle{tracker: tracker}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if tracker.Open() != 0 || tracker.Maximum() > 1 {
		t.Fatalf("resource counts open=%d max=%d", tracker.Open(), tracker.Maximum())
	}
}

func TestPooledAuditDoesNotLeakPreviousIdentity(t *testing.T) {
	memory := store.NewMemory()
	runtime := New(memory)
	runtime.AuditRequest("user-a", "campaign-a", "trace-secret")
	runtime.AuditRequest("user-b", "campaign-b", "")
	audits := memory.Audits()
	if len(audits) != 2 {
		t.Fatalf("audit count = %d", len(audits))
	}
	if audits[0].UserID != "user-a" || audits[0].Correlation != "trace-secret" {
		t.Fatalf("first audit corrupted: %#v", audits[0])
	}
	if audits[1].UserID != "user-b" || audits[1].Correlation != "" {
		t.Fatalf("second audit leaked identity: %#v", audits[1])
	}
}

func TestConcurrentLastPrizeProducesOneWinner(t *testing.T) {
	memory := store.NewMemory()
	memory.PutPrize(&model.PrizeSnapshot{PrizeID: "last-prize", Remaining: 1})
	runtime := New(memory)
	ready := make(chan struct{}, 2)
	start := make(chan struct{})
	results := make(chan bool, 2)
	for i := 0; i < 2; i++ {
		go func() { results <- runtime.ClaimPrize("last-prize", ready, start) }()
	}
	<-ready
	<-ready
	close(start)
	wins := 0
	for i := 0; i < 2; i++ {
		if <-results {
			wins++
		}
	}
	if wins != 1 {
		t.Fatalf("winner count = %d", wins)
	}
	if remaining := memory.Prize("last-prize").Remaining; remaining != 0 {
		t.Fatalf("remaining = %d", remaining)
	}
}

func TestBatchDrawReturnsEveryTaskBeforeClosing(t *testing.T) {
	runtime := New(store.NewMemory())
	tasks := []model.DrawTask{{ID: "draw-a"}, {ID: "draw-b"}, {ID: "draw-c"}}
	start := make(chan struct{})
	var got []string
	done := make(chan struct{})
	go func() {
		got = runtime.BatchDraw(tasks, start)
		close(done)
	}()
	close(start)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("batch draw did not close")
	}
	want := map[string]bool{"draw-a": true, "draw-b": true, "draw-c": true}
	if len(got) != len(want) {
		t.Fatalf("result count = %d: %v", len(got), got)
	}
	for _, id := range got {
		if !want[id] {
			t.Fatalf("unexpected result %q", id)
		}
	}
}

func TestCancelledDeliveryStopsRetriesAndNextRequestIsClean(t *testing.T) {
	memory := store.NewMemory()
	runtime := New(memory)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	retry := make(chan struct{})
	done := runtime.Dispatch(ctx, "campaign-old", started, retry, func() error { return errors.New("offline") })
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled delivery returned %v", err)
	}
	old := memory.Delivery("campaign-old")
	if old == nil || !old.Stopped || old.Attempts != 1 {
		t.Fatalf("cancelled delivery state = %#v", old)
	}
	started2 := make(chan struct{})
	retry2 := make(chan struct{})
	done2 := runtime.Dispatch(context.Background(), "campaign-new", started2, retry2, func() error { return nil })
	<-started2
	if err := <-done2; err != nil {
		t.Fatalf("next delivery inherited cancellation: %v", err)
	}
	if next := memory.Delivery("campaign-new"); next == nil || next.Stopped || next.Attempts != 1 {
		t.Fatalf("next delivery state = %#v", next)
	}
}

func TestFailedClaimCommitDoesNotPublishAudit(t *testing.T) {
	memory := store.NewMemory()
	memory.PutClaim(&model.Claim{ID: "claim-1", PrizeID: "prize-1", State: "won"})
	memory.SetCommitError(errors.New("storage unavailable"))
	runtime := New(memory)
	if err := runtime.CompleteClaim("claim-1"); err == nil {
		t.Fatal("expected commit failure")
	}
	if claim := memory.Claim("claim-1"); claim == nil || claim.State != "won" {
		t.Fatalf("failed claim changed persisted state: %#v", claim)
	}
	if audits := memory.ClaimAudits(); len(audits) != 0 {
		t.Fatalf("failed claim published audit: %v", audits)
	}
	if err := runtime.CompleteClaim("claim-1"); err != nil {
		t.Fatalf("retry claim failed: %v", err)
	}
	if claim := memory.Claim("claim-1"); claim.State != "claimed" {
		t.Fatalf("retried claim state = %#v", claim)
	}
	if audits := memory.ClaimAudits(); !reflect.DeepEqual(audits, []string{"claim-1"}) {
		t.Fatalf("retried claim audits = %v", audits)
	}
}

func TestRuntimeFlowHasNoUnexpectedRaces(t *testing.T) {
	memory := store.NewMemory()
	runtime := New(memory)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			runtime.AuditRequest(fmt.Sprintf("u-%d", i), "campaign", "")
		}(i)
	}
	wg.Wait()
}
