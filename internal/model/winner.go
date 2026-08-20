package model

import (
	"time"
)

const (
	WinnerStatusWon    = "won"
	WinnerStatusClaimed = "claimed"
	WinnerStatusExpired = "expired"
)

var winnerTransitions = map[string]map[string]bool{
	WinnerStatusWon: {WinnerStatusClaimed: true, WinnerStatusExpired: true},
}

func WinnerCanTransition(from, to string) bool {
	if m, ok := winnerTransitions[from]; ok {
		return m[to]
	}
	return false
}

type Winner struct {
	ID        string    `json:"id"`
	ActivityID string   `json:"activity_id"`
	PrizeID   string    `json:"prize_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	WonAt     time.Time `json:"won_at"`
	ClaimedAt time.Time `json:"claimed_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (w *Winner) Validate() error {
	if w.ActivityID == "" {
		return NewValidationError("activity_id", "活动 ID 不能为空")
	}
	if w.PrizeID == "" {
		return NewValidationError("prize_id", "奖品 ID 不能为空")
	}
	if w.UserID == "" {
		return NewValidationError("user_id", "用户 ID 不能为空")
	}
	if w.Status == "" {
		w.Status = WinnerStatusWon
	}
	if w.Status != WinnerStatusWon && w.Status != WinnerStatusClaimed && w.Status != WinnerStatusExpired {
		return NewValidationError("status", "中奖状态不合法")
	}
	return nil
}

type WinnerFilter struct {
	ActivityID string
	UserID     string
	Status     string
}

func (f WinnerFilter) Match(w *Winner) bool {
	if f.ActivityID != "" && w.ActivityID != f.ActivityID {
		return false
	}
	if f.UserID != "" && w.UserID != f.UserID {
		return false
	}
	if f.Status != "" && w.Status != f.Status {
		return false
	}
	return true
}
