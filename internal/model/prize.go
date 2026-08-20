package model

import (
	"strings"
	"time"
)

type Prize struct {
	ID        string    `json:"id"`
	ActivityID string   `json:"activity_id"`
	Name      string    `json:"name"`
	Total     int       `json:"total"`
	Remaining int       `json:"remaining"`
	Weight    int       `json:"weight"`
	Level     int       `json:"level"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *Prize) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return NewValidationError("name", "奖品名称不能为空")
	}
	if p.ActivityID == "" {
		return NewValidationError("activity_id", "活动 ID 不能为空")
	}
	if p.Total < 0 {
		return NewValidationError("total", "库存总量不能为负数")
	}
	if p.Remaining < 0 {
		return NewValidationError("remaining", "剩余数量不能为负数")
	}
	if p.Weight < 0 {
		return NewValidationError("weight", "权重不能为负数")
	}
	if p.Level < 1 {
		return NewValidationError("level", "奖品等级必须大于 0")
	}
	return nil
}

type PrizeFilter struct {
	ActivityID string
	Keyword    string
}

func (f PrizeFilter) Match(p *Prize) bool {
	if f.ActivityID != "" && p.ActivityID != f.ActivityID {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(p.Name), k) {
			return false
		}
	}
	return true
}
