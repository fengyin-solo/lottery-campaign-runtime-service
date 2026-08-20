package store

import (
	"lottery/internal/model"
)

func (s *MemoryStore) CreatePrize(p *model.Prize) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prizes[p.ID] = p
	return nil
}

func (s *MemoryStore) GetPrize(id string) (*model.Prize, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.prizes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStore) ListPrizes() []*model.Prize {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Prize, 0, len(s.prizes))
	for _, p := range s.prizes {
		list = append(list, p)
	}
	return list
}

func (s *MemoryStore) UpdatePrize(p *model.Prize) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.prizes[p.ID]; !ok {
		return ErrNotFound
	}
	s.prizes[p.ID] = p
	return nil
}

func (s *MemoryStore) DeletePrize(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.prizes[id]; !ok {
		return ErrNotFound
	}
	delete(s.prizes, id)
	return nil
}

func (s *MemoryStore) DecrementPrizeRemaining(prizeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.prizes[prizeID]
	if !ok {
		return ErrNotFound
	}
	if p.Remaining <= 0 {
		return ErrConflict
	}
	p.Remaining--
	return nil
}
