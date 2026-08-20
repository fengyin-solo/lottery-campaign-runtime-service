package model

import (
	"time"
)

const (
	EntryResultWin    = "win"
	EntryResultLose   = "lose"
)

type Entry struct {
	ID         string    `json:"id"`
	ActivityID string    `json:"activity_id"`
	UserID     string    `json:"user_id"`
	Result     string    `json:"result"`
	CreatedAt  time.Time `json:"created_at"`
}

func (e *Entry) Validate() error {
	if e.ActivityID == "" {
		return NewValidationError("activity_id", "活动 ID 不能为空")
	}
	if e.UserID == "" {
		return NewValidationError("user_id", "用户 ID 不能为空")
	}
	if e.Result == "" {
		e.Result = EntryResultLose
	}
	if e.Result != EntryResultWin && e.Result != EntryResultLose {
		return NewValidationError("result", "结果类型不合法")
	}
	return nil
}

type EntryFilter struct {
	ActivityID string
	UserID     string
	Result     string
}

func (f EntryFilter) Match(e *Entry) bool {
	if f.ActivityID != "" && e.ActivityID != f.ActivityID {
		return false
	}
	if f.UserID != "" && e.UserID != f.UserID {
		return false
	}
	if f.Result != "" && e.Result != f.Result {
		return false
	}
	return true
}
