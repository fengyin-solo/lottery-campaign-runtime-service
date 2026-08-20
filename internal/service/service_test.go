package service

import (
	"testing"

	"lottery/internal/config"
	"lottery/internal/model"
	"lottery/internal/store"
	"lottery/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func TestActivityLifecycle(t *testing.T) {
	svc := newTestService()
	a, err := svc.CreateActivity(model.Activity{Name: "test", Status: model.ActivityDraft, LimitPerUser: 3})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if a.Status != model.ActivityDraft {
		t.Errorf("default status should be draft")
	}
	_, err = svc.TransitionActivityStatus(a.ID, model.ActivityActive)
	if err != nil {
		t.Fatalf("transition to active: %v", err)
	}
	_, err = svc.TransitionActivityStatus(a.ID, model.ActivityDraft)
	if err == nil {
		t.Errorf("expected error for illegal transition")
	}
	_, err = svc.TransitionActivityStatus(a.ID, model.ActivityEnded)
	if err != nil {
		t.Fatalf("transition to ended: %v", err)
	}
	_, err = svc.UpdateActivity(a.ID, model.Activity{Name: "updated", Status: model.ActivityEnded})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestPrizeCRUD(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityDraft})
	p, err := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "prize", Total: 10, Weight: 5, Level: 1})
	if err != nil {
		t.Fatalf("create prize: %v", err)
	}
	if p.Remaining != 10 {
		t.Errorf("remaining should be 10")
	}
	_, err = svc.CreatePrize(model.Prize{ActivityID: "bad", Name: "x", Total: 1, Weight: 1, Level: 1})
	if err == nil {
		t.Errorf("expected error for missing activity")
	}
	up, err := svc.UpdatePrize(p.ID, model.Prize{Name: "new", Total: 20, Weight: 5, Level: 1})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if up.Remaining != 20 {
		t.Errorf("remaining should update to 20")
	}
}

func TestDrawAndLimit(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive, LimitPerUser: 2})
	prize, _ := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "gold", Total: 5, Weight: 100, Level: 1})
	user, _ := svc.CreateUser(model.User{Name: "alice"})

	res1, err := svc.Draw(act.ID, user.ID)
	if err != nil {
		t.Fatalf("draw 1: %v", err)
	}
	if res1.Entry == nil {
		t.Errorf("entry nil")
	}
	p, _ := svc.GetPrize(prize.ID)
	if p.Remaining != 4 {
		t.Errorf("remaining should be 4, got %d", p.Remaining)
	}

	_, err = svc.Draw(act.ID, user.ID)
	if err != nil {
		t.Fatalf("draw 2: %v", err)
	}
	_, err = svc.Draw(act.ID, user.ID)
	if err == nil {
		t.Errorf("expected limit exceeded error")
	}

	act2, _ := svc.CreateActivity(model.Activity{Name: "act2", Status: model.ActivityDraft, LimitPerUser: 10})
	_, err = svc.Draw(act2.ID, user.ID)
	if err == nil {
		t.Errorf("expected error for non-active activity")
	}
}

func TestBatchDraw(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive, LimitPerUser: 10})
	_, _ = svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p1", Total: 100, Weight: 100, Level: 1})
	user, _ := svc.CreateUser(model.User{Name: "bob"})

	results, err := svc.BatchDraw(act.ID, user.ID, 3)
	if err != nil {
		t.Fatalf("batch draw: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results")
	}
	_, err = svc.BatchDraw(act.ID, user.ID, 0)
	if err == nil {
		t.Errorf("expected error for zero count")
	}
}

func TestWinnerClaim(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive})
	prize, _ := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 10, Weight: 100, Level: 1})
	user, _ := svc.CreateUser(model.User{Name: "cathy"})

	w, err := svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: user.ID})
	if err != nil {
		t.Fatalf("create winner: %v", err)
	}
	claimed, err := svc.ClaimWinner(w.ID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if claimed.Status != model.WinnerStatusClaimed {
		t.Errorf("status should be claimed")
	}
	_, err = svc.ClaimWinner(w.ID)
	if err == nil {
		t.Errorf("expected error for double claim")
	}
	expired, _ := svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: user.ID})
	_, err = svc.UpdateWinner(expired.ID, model.Winner{Status: model.WinnerStatusExpired})
	if err != nil {
		t.Fatalf("expire: %v", err)
	}
	cnt, err := svc.BatchExpireWinners()
	if err != nil {
		t.Fatalf("batch expire: %v", err)
	}
	if cnt < 0 {
		t.Errorf("unexpected count")
	}
}

func TestStats(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive})
	prize, _ := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 10, Weight: 100, Level: 1})
	user, _ := svc.CreateUser(model.User{Name: "dave"})

	_, _ = svc.Draw(act.ID, user.ID)
	_, _ = svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: user.ID})

	stats, err := svc.GetActivityStats(act.ID)
	if err != nil {
		t.Fatalf("activity stats: %v", err)
	}
	if stats.EntryCount < 1 {
		t.Errorf("entry count should >= 1")
	}
	if stats.WinnerCount < 1 {
		t.Errorf("winner count should >= 1")
	}

	ps, err := svc.GetPrizeStats(act.ID)
	if err != nil {
		t.Fatalf("prize stats: %v", err)
	}
	if len(ps) < 1 {
		t.Errorf("expected prize stats")
	}
}

func TestUserCRUD(t *testing.T) {
	svc := newTestService()
	u, err := svc.CreateUser(model.User{Name: "eve"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	got, err := svc.GetUser(u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.Name != "eve" {
		t.Errorf("name mismatch")
	}
	_, err = svc.CreateUser(model.User{Name: "eve"})
	if err == nil {
		t.Errorf("expected conflict")
	}
	_, err = svc.UpdateUser(u.ID, model.User{Name: "frank"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := svc.DeleteUser(u.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestEntryAndWinnerListFilter(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive})
	prize, _ := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 10, Weight: 1, Level: 1})
	u1, _ := svc.CreateUser(model.User{Name: "g1"})
	u2, _ := svc.CreateUser(model.User{Name: "g2"})

	_, _ = svc.CreateEntry(model.Entry{ActivityID: act.ID, UserID: u1.ID, Result: model.EntryResultWin})
	_, _ = svc.CreateEntry(model.Entry{ActivityID: act.ID, UserID: u2.ID, Result: model.EntryResultLose})
	_, _ = svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: u1.ID})

	entries, total, _ := svc.ListEntries(model.EntryFilter{ActivityID: act.ID, UserID: u1.ID}, 1, 10)
	if total != 1 || len(entries) != 1 {
		t.Errorf("entry filter mismatch")
	}
	winners, total, _ := svc.ListWinners(model.WinnerFilter{ActivityID: act.ID, Status: model.WinnerStatusWon}, 1, 10)
	if total < 1 {
		t.Errorf("winner filter mismatch")
	}
	if len(winners) < 1 {
		t.Errorf("expected winners")
	}
}

func TestActivityListFilter(t *testing.T) {
	svc := newTestService()
	_, _ = svc.CreateActivity(model.Activity{Name: "alpha", Status: model.ActivityDraft})
	_, _ = svc.CreateActivity(model.Activity{Name: "beta", Status: model.ActivityActive})
	items, total, _ := svc.ListActivities(model.ActivityFilter{Status: model.ActivityActive}, 1, 10)
	if total != 1 || len(items) != 1 || items[0].Name != "beta" {
		t.Errorf("activity filter mismatch")
	}
	items, total, _ = svc.ListActivities(model.ActivityFilter{Keyword: "alp"}, 1, 10)
	if total != 1 || items[0].Name != "alpha" {
		t.Errorf("keyword filter mismatch")
	}
}

func TestDrawConcurrentStock(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive, LimitPerUser: 1000})
	_, _ = svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 10, Weight: 100, Level: 1})
	user, _ := svc.CreateUser(model.User{Name: "concurrent"})

	for i := 0; i < 20; i++ {
		_, _ = svc.Draw(act.ID, user.ID)
	}
	prizes, _, _ := svc.ListPrizes(model.PrizeFilter{ActivityID: act.ID}, 1, 10)
	if len(prizes) > 0 && prizes[0].Remaining < 0 {
		t.Errorf("remaining should never be negative")
	}
}

func TestActivityPagination(t *testing.T) {
	svc := newTestService()
	for i := 0; i < 5; i++ {
		_, _ = svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityDraft})
	}
	items, total, err := svc.ListActivities(model.ActivityFilter{}, 1, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 5 || len(items) != 2 {
		t.Errorf("expected total 5 page 2, got %d %d", total, len(items))
	}
	items, total, err = svc.ListActivities(model.ActivityFilter{}, 3, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected last page 1 item, got %d", len(items))
	}
}

func TestPrizePagination(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityDraft})
	for i := 0; i < 5; i++ {
		_, _ = svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 1, Weight: 1, Level: i + 1})
	}
	items, total, err := svc.ListPrizes(model.PrizeFilter{ActivityID: act.ID}, 1, 3)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 5 || len(items) != 3 {
		t.Errorf("expected total 5 page 3, got %d %d", total, len(items))
	}
}

func TestWinnerPaginationAndFilter(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive})
	prize, _ := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 10, Weight: 1, Level: 1})
	u1, _ := svc.CreateUser(model.User{Name: "w1"})
	u2, _ := svc.CreateUser(model.User{Name: "w2"})
	_, _ = svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: u1.ID})
	_, _ = svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: u2.ID})

	items, total, err := svc.ListWinners(model.WinnerFilter{ActivityID: act.ID}, 1, 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(items) != 1 {
		t.Errorf("expected 2 total 1 page, got %d %d", total, len(items))
	}
	items, total, _ = svc.ListWinners(model.WinnerFilter{UserID: u1.ID}, 1, 10)
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 for user filter")
	}
}

func TestUserPagination(t *testing.T) {
	svc := newTestService()
	for i := 0; i < 5; i++ {
		_, _ = svc.CreateUser(model.User{Name: "user" + string(rune('a'+i))})
	}
	items, total, err := svc.ListUsers(model.UserFilter{}, 2, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 5 || len(items) != 2 {
		t.Errorf("expected total 5 page 2, got %d %d", total, len(items))
	}
}

func TestBatchExpireWinners(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive})
	prize, _ := svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "p", Total: 10, Weight: 1, Level: 1})
	u1, _ := svc.CreateUser(model.User{Name: "e1"})
	u2, _ := svc.CreateUser(model.User{Name: "e2"})
	w1, _ := svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: u1.ID})
	_, _ = svc.CreateWinner(model.Winner{ActivityID: act.ID, PrizeID: prize.ID, UserID: u2.ID})
	_, _ = svc.ClaimWinner(w1.ID)

	cnt, err := svc.BatchExpireWinners()
	if err != nil {
		t.Fatalf("batch expire: %v", err)
	}
	if cnt != 1 {
		t.Errorf("expected 1 expired, got %d", cnt)
	}
	wins, _, _ := svc.ListWinners(model.WinnerFilter{Status: model.WinnerStatusExpired}, 1, 10)
	if len(wins) != 1 {
		t.Errorf("expected 1 expired winner")
	}
}

func TestDrawNoPrize(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive, LimitPerUser: 10})
	user, _ := svc.CreateUser(model.User{Name: "noprize"})
	res, err := svc.Draw(act.ID, user.ID)
	if err != nil {
		t.Fatalf("draw: %v", err)
	}
	if res.Entry == nil {
		t.Errorf("entry should exist")
	}
	if res.Winner != nil {
		t.Errorf("winner should be nil when no prize")
	}
}

func TestDrawWeightedRandom(t *testing.T) {
	svc := newTestService()
	act, _ := svc.CreateActivity(model.Activity{Name: "act", Status: model.ActivityActive, LimitPerUser: 100})
	_, _ = svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "high", Total: 100, Weight: 100, Level: 1})
	_, _ = svc.CreatePrize(model.Prize{ActivityID: act.ID, Name: "low", Total: 100, Weight: 1, Level: 2})
	user, _ := svc.CreateUser(model.User{Name: "weighted"})
	winCounts := make(map[string]int)
	for i := 0; i < 50; i++ {
		res, err := svc.Draw(act.ID, user.ID)
		if err != nil {
			break
		}
		if res.Winner != nil {
			winCounts[res.Prize.Name]++
		}
	}
	if winCounts["high"] <= winCounts["low"] {
		t.Errorf("high weight prize should win more often")
	}
}
