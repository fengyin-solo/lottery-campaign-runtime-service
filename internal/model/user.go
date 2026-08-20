package model

import (
	"strings"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	EntryCount int      `json:"entry_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return NewValidationError("name", "用户名称不能为空")
	}
	if u.EntryCount < 0 {
		return NewValidationError("entry_count", "参与次数不能为负数")
	}
	return nil
}

type UserFilter struct {
	Keyword string
}

func (f UserFilter) Match(u *User) bool {
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(u.Name), k) {
			return false
		}
	}
	return true
}
