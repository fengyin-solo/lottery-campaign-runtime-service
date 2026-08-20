package store

import (
	"lottery/internal/model"
)

func (s *MemoryStore) CreateActivity(a *model.Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activities[a.ID] = a
	return nil
}

func (s *MemoryStore) GetActivity(id string) (*model.Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.activities[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) ListActivities() []*model.Activity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Activity, 0, len(s.activities))
	for _, a := range s.activities {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateActivity(a *model.Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.activities[a.ID]; !ok {
		return ErrNotFound
	}
	s.activities[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteActivity(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.activities[id]; !ok {
		return ErrNotFound
	}
	delete(s.activities, id)
	return nil
}
