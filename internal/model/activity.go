package model

import (
	"strings"
	"time"
)

const (
	ActivityDraft   = "draft"
	ActivityActive  = "active"
	ActivityEnded   = "ended"
)

var activityTransitions = map[string]map[string]bool{
	ActivityDraft:  {ActivityActive: true},
	ActivityActive: {ActivityEnded: true},
}

func ActivityCanTransition(from, to string) bool {
	if m, ok := activityTransitions[from]; ok {
		return m[to]
	}
	return false
}

type Activity struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	LimitPerUser int       `json:"limit_per_user"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (a *Activity) Validate() error {
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return NewValidationError("name", "活动名称不能为空")
	}
	if a.Status == "" {
		a.Status = ActivityDraft
	}
	if a.Status != ActivityDraft && a.Status != ActivityActive && a.Status != ActivityEnded {
		return NewValidationError("status", "活动状态不合法")
	}
	if !a.StartTime.IsZero() && !a.EndTime.IsZero() && a.EndTime.Before(a.StartTime) {
		return NewValidationError("end_time", "结束时间不能早于开始时间")
	}
	if a.LimitPerUser < 0 {
		return NewValidationError("limit_per_user", "每人限抽次数不能为负数")
	}
	return nil
}

type ActivityFilter struct {
	Status  string
	Keyword string
}

func (f ActivityFilter) Match(a *Activity) bool {
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(a.Name), k) &&
			!strings.Contains(strings.ToLower(a.Description), k) {
			return false
		}
	}
	return true
}
