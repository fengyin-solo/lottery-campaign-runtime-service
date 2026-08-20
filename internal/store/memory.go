package store

import (
	"sync"

	"lottery/internal/model"
)

type MemoryStore struct {
	mu        sync.RWMutex
	activities map[string]*model.Activity
	prizes     map[string]*model.Prize
	entries    map[string]*model.Entry
	winners    map[string]*model.Winner
	users      map[string]*model.User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		activities: make(map[string]*model.Activity),
		prizes:     make(map[string]*model.Prize),
		entries:    make(map[string]*model.Entry),
		winners:    make(map[string]*model.Winner),
		users:      make(map[string]*model.User),
	}
}

var _ Store = (*MemoryStore)(nil)
