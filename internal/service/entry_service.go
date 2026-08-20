package service

import (
	"sort"
	"time"

	"lottery/internal/model"
	"lottery/pkg/idgen"
)

func (s *Service) CreateEntry(input model.Entry) (*model.Entry, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetActivity(input.ActivityID); err != nil {
		return nil, model.NewValidationError("activity_id", "关联活动不存在")
	}
	e := &model.Entry{
		ID:         idgen.Hex(),
		ActivityID: input.ActivityID,
		UserID:     input.UserID,
		Result:     input.Result,
		CreatedAt:  time.Now(),
	}
	if err := s.store.CreateEntry(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) GetEntry(id string) (*model.Entry, error) {
	return s.store.GetEntry(id)
}

func (s *Service) ListEntries(filter model.EntryFilter, page, size int) ([]*model.Entry, int, error) {
	all := s.store.ListEntries()
	matched := make([]*model.Entry, 0, len(all))
	for _, e := range all {
		if filter.Match(e) {
			matched = append(matched, e)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Entry{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateEntry(id string, input model.Entry) (*model.Entry, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	e, err := s.store.GetEntry(id)
	if err != nil {
		return nil, err
	}
	e.Result = input.Result
	if err := s.store.UpdateEntry(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) DeleteEntry(id string) error {
	return s.store.DeleteEntry(id)
}
