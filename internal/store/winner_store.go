package store

import (
	"lottery/internal/model"
)

func (s *MemoryStore) CreateWinner(w *model.Winner) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.winners[w.ID] = w
	return nil
}

func (s *MemoryStore) GetWinner(id string) (*model.Winner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.winners[id]
	if !ok {
		return nil, ErrNotFound
	}
	return w, nil
}

func (s *MemoryStore) ListWinners() []*model.Winner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Winner, 0, len(s.winners))
	for _, w := range s.winners {
		list = append(list, w)
	}
	return list
}

func (s *MemoryStore) UpdateWinner(w *model.Winner) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.winners[w.ID]; !ok {
		return ErrNotFound
	}
	s.winners[w.ID] = w
	return nil
}

func (s *MemoryStore) DeleteWinner(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.winners[id]; !ok {
		return ErrNotFound
	}
	delete(s.winners, id)
	return nil
}
