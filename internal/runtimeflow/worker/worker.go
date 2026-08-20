package worker

import (
	"context"
	"errors"
	"sync"

	"lottery/internal/runtimeflow/model"
)

func RunExport(ctx context.Context, started chan<- struct{}, release <-chan struct{}) error {
	close(started)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-release:
		return nil
	}
}

func ConsumeAudience(batch model.AudienceBatch, started chan<- struct{}, release <-chan struct{}) []string {
	owned := batch.Clone()
	close(started)
	<-release
	return append([]string(nil), owned.UserIDs...)
}

func RetryRedemption(ctx context.Context, max int, attempt func(int) error) (int, error) {
	for i := 1; i <= max; i++ {
		if err := ctx.Err(); err != nil {
			return i - 1, err
		}
		err := attempt(i)
		if err == nil {
			return i, nil
		}
		if !errors.Is(err, model.ErrTemporary) {
			return i, err
		}
	}
	return max, model.ErrTemporary
}

func Notify(ctx context.Context, notifier model.Notifier, userID string) error {
	if notifier == nil {
		return nil
	}
	return notifier.Notify(ctx, userID)
}

type BatchHandle interface {
	Process(string) error
	Close() error
}

func ProcessResources(ids []string, open func() (BatchHandle, error)) error {
	for _, id := range ids {
		handle, err := open()
		if err != nil {
			return err
		}
		if err := handle.Process(id); err != nil {
			_ = handle.Close()
			return err
		}
		if err := handle.Close(); err != nil {
			return err
		}
	}
	return nil
}

func FanOut(tasks []model.DrawTask, start <-chan struct{}) <-chan string {
	results := make(chan string, len(tasks))
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(task model.DrawTask) {
			defer wg.Done()
			<-start
			task.MarkDelivered()
			results <- task.ID
		}(task.Clone())
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

func SnapshotAvailable(snapshot *model.PrizeSnapshot) bool {
	return snapshot != nil && snapshot.Remaining > 0
}

func CommitSucceeded(err error) bool { return err == nil }

func DeliverWithRetry(ctx context.Context, started chan<- struct{}, retry <-chan struct{}, send func() error) (int, error) {
	close(started)
	attempts := 0
	for {
		if err := ctx.Err(); err != nil {
			return attempts, err
		}
		attempts++
		if err := send(); err == nil {
			return attempts, nil
		}
		select {
		case <-ctx.Done():
			return attempts, ctx.Err()
		case <-retry:
		}
	}
}
