// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"lottery/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// Activity
	CreateActivity(a *model.Activity) error
	GetActivity(id string) (*model.Activity, error)
	ListActivities() []*model.Activity
	UpdateActivity(a *model.Activity) error
	DeleteActivity(id string) error

	// Prize
	CreatePrize(p *model.Prize) error
	GetPrize(id string) (*model.Prize, error)
	ListPrizes() []*model.Prize
	UpdatePrize(p *model.Prize) error
	DeletePrize(id string) error

	// Entry
	CreateEntry(e *model.Entry) error
	GetEntry(id string) (*model.Entry, error)
	ListEntries() []*model.Entry
	UpdateEntry(e *model.Entry) error
	DeleteEntry(id string) error

	// Winner
	CreateWinner(w *model.Winner) error
	GetWinner(id string) (*model.Winner, error)
	ListWinners() []*model.Winner
	UpdateWinner(w *model.Winner) error
	DeleteWinner(id string) error

	// User
	CreateUser(u *model.User) error
	GetUser(id string) (*model.User, error)
	GetUserByName(name string) (*model.User, error)
	ListUsers() []*model.User
	UpdateUser(u *model.User) error
	DeleteUser(id string) error

	// Lottery atomic operation
	DecrementPrizeRemaining(prizeID string) error
}
