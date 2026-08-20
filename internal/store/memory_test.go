package store

import (
	"testing"
	"time"

	"lottery/internal/model"
)

func TestMemoryStoreActivity(t *testing.T) {
	s := NewMemoryStore()
	a := &model.Activity{ID: "a1", Name: "test", Status: model.ActivityDraft, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateActivity(a); err != nil {
		t.Fatalf("create activity: %v", err)
	}
	got, err := s.GetActivity("a1")
	if err != nil {
		t.Fatalf("get activity: %v", err)
	}
	if got.Name != "test" {
		t.Errorf("name mismatch")
	}
	if len(s.ListActivities()) != 1 {
		t.Errorf("list count mismatch")
	}
	a.Name = "updated"
	if err := s.UpdateActivity(a); err != nil {
		t.Fatalf("update activity: %v", err)
	}
	if err := s.DeleteActivity("a1"); err != nil {
		t.Fatalf("delete activity: %v", err)
	}
	if _, err := s.GetActivity("a1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete")
	}
}

func TestMemoryStorePrize(t *testing.T) {
	s := NewMemoryStore()
	p := &model.Prize{ID: "p1", ActivityID: "a1", Name: "prize", Total: 10, Remaining: 10, Weight: 5, Level: 1, CreatedAt: time.Now()}
	if err := s.CreatePrize(p); err != nil {
		t.Fatalf("create prize: %v", err)
	}
	got, err := s.GetPrize("p1")
	if err != nil {
		t.Fatalf("get prize: %v", err)
	}
	if got.Name != "prize" {
		t.Errorf("name mismatch")
	}
	if err := s.DecrementPrizeRemaining("p1"); err != nil {
		t.Fatalf("decrement: %v", err)
	}
	got, _ = s.GetPrize("p1")
	if got.Remaining != 9 {
		t.Errorf("remaining should be 9, got %d", got.Remaining)
	}
	for i := 0; i < 9; i++ {
		_ = s.DecrementPrizeRemaining("p1")
	}
	if err := s.DecrementPrizeRemaining("p1"); err != ErrConflict {
		t.Errorf("expected ErrConflict when empty, got %v", err)
	}
	if err := s.DeletePrize("p1"); err != nil {
		t.Fatalf("delete prize: %v", err)
	}
	if _, err := s.GetPrize("p1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete")
	}
}

func TestMemoryStoreEntry(t *testing.T) {
	s := NewMemoryStore()
	e := &model.Entry{ID: "e1", ActivityID: "a1", UserID: "u1", Result: model.EntryResultWin, CreatedAt: time.Now()}
	if err := s.CreateEntry(e); err != nil {
		t.Fatalf("create entry: %v", err)
	}
	got, err := s.GetEntry("e1")
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if got.UserID != "u1" {
		t.Errorf("user id mismatch")
	}
	if len(s.ListEntries()) != 1 {
		t.Errorf("list count mismatch")
	}
	if err := s.DeleteEntry("e1"); err != nil {
		t.Fatalf("delete entry: %v", err)
	}
}

func TestMemoryStoreWinner(t *testing.T) {
	s := NewMemoryStore()
	w := &model.Winner{ID: "w1", ActivityID: "a1", PrizeID: "p1", UserID: "u1", Status: model.WinnerStatusWon, WonAt: time.Now(), CreatedAt: time.Now()}
	if err := s.CreateWinner(w); err != nil {
		t.Fatalf("create winner: %v", err)
	}
	got, err := s.GetWinner("w1")
	if err != nil {
		t.Fatalf("get winner: %v", err)
	}
	if got.Status != model.WinnerStatusWon {
		t.Errorf("status mismatch")
	}
	w.Status = model.WinnerStatusClaimed
	if err := s.UpdateWinner(w); err != nil {
		t.Fatalf("update winner: %v", err)
	}
	if err := s.DeleteWinner("w1"); err != nil {
		t.Fatalf("delete winner: %v", err)
	}
	if _, err := s.GetWinner("w1"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound")
	}
}

func TestMemoryStoreUser(t *testing.T) {
	s := NewMemoryStore()
	u := &model.User{ID: "u1", Name: "alice", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateUser(u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := s.CreateUser(&model.User{ID: "u2", Name: "alice"}); err != ErrConflict {
		t.Errorf("expected conflict on same name")
	}
	got, err := s.GetUser("u1")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.Name != "alice" {
		t.Errorf("name mismatch")
	}
	byName, err := s.GetUserByName("alice")
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}
	if byName.ID != "u1" {
		t.Errorf("by name id mismatch")
	}
	u.Name = "bob"
	if err := s.UpdateUser(u); err != nil {
		t.Fatalf("update user: %v", err)
	}
	if err := s.DeleteUser("u1"); err != nil {
		t.Fatalf("delete user: %v", err)
	}
}

func TestMemoryStoreNotFound(t *testing.T) {
	s := NewMemoryStore()
	cases := []struct {
		name string
		f    func() error
	}{
		{"activity", func() error { _, err := s.GetActivity("x"); return err }},
		{"prize", func() error { _, err := s.GetPrize("x"); return err }},
		{"entry", func() error { _, err := s.GetEntry("x"); return err }},
		{"winner", func() error { _, err := s.GetWinner("x"); return err }},
		{"user", func() error { _, err := s.GetUser("x"); return err }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.f(); err != ErrNotFound {
				t.Errorf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

func TestMemoryStoreUpdateNotFound(t *testing.T) {
	s := NewMemoryStore()
	if err := s.UpdateActivity(&model.Activity{ID: "x"}); err != ErrNotFound {
		t.Errorf("activity update: expected ErrNotFound, got %v", err)
	}
	if err := s.UpdatePrize(&model.Prize{ID: "x"}); err != ErrNotFound {
		t.Errorf("prize update: expected ErrNotFound, got %v", err)
	}
	if err := s.UpdateEntry(&model.Entry{ID: "x"}); err != ErrNotFound {
		t.Errorf("entry update: expected ErrNotFound, got %v", err)
	}
	if err := s.UpdateWinner(&model.Winner{ID: "x"}); err != ErrNotFound {
		t.Errorf("winner update: expected ErrNotFound, got %v", err)
	}
	if err := s.UpdateUser(&model.User{ID: "x"}); err != ErrNotFound {
		t.Errorf("user update: expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStoreDeleteNotFound(t *testing.T) {
	s := NewMemoryStore()
	if err := s.DeleteActivity("x"); err != ErrNotFound {
		t.Errorf("activity delete: expected ErrNotFound, got %v", err)
	}
	if err := s.DeletePrize("x"); err != ErrNotFound {
		t.Errorf("prize delete: expected ErrNotFound, got %v", err)
	}
	if err := s.DeleteEntry("x"); err != ErrNotFound {
		t.Errorf("entry delete: expected ErrNotFound, got %v", err)
	}
	if err := s.DeleteWinner("x"); err != ErrNotFound {
		t.Errorf("winner delete: expected ErrNotFound, got %v", err)
	}
	if err := s.DeleteUser("x"); err != ErrNotFound {
		t.Errorf("user delete: expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStoreDecrementNotFound(t *testing.T) {
	s := NewMemoryStore()
	if err := s.DecrementPrizeRemaining("x"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
