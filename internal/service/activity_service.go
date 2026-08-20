package service

import (
	"sort"
	"time"

	"lottery/internal/model"
	"lottery/pkg/idgen"
)

func (s *Service) CreateActivity(input model.Activity) (*model.Activity, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	a := &model.Activity{
		ID:           idgen.Hex(),
		Name:         input.Name,
		Description:  input.Description,
		StartTime:    input.StartTime,
		EndTime:      input.EndTime,
		Status:       input.Status,
		LimitPerUser: input.LimitPerUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if a.Status == "" {
		a.Status = model.ActivityDraft
	}
	if err := s.store.CreateActivity(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) GetActivity(id string) (*model.Activity, error) {
	return s.store.GetActivity(id)
}

func (s *Service) ListActivities(filter model.ActivityFilter, page, size int) ([]*model.Activity, int, error) {
	all := s.store.ListActivities()
	matched := make([]*model.Activity, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Activity{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateActivity(id string, input model.Activity) (*model.Activity, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	a, err := s.store.GetActivity(id)
	if err != nil {
		return nil, err
	}
	if input.Status != "" && input.Status != a.Status {
		if !model.ActivityCanTransition(a.Status, input.Status) {
			return nil, model.NewValidationError("status", "活动状态流转不合法")
		}
		a.Status = input.Status
	}
	a.Name = input.Name
	a.Description = input.Description
	a.StartTime = input.StartTime
	a.EndTime = input.EndTime
	a.LimitPerUser = input.LimitPerUser
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateActivity(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) DeleteActivity(id string) error {
	return s.store.DeleteActivity(id)
}

func (s *Service) TransitionActivityStatus(id string, toStatus string) (*model.Activity, error) {
	a, err := s.store.GetActivity(id)
	if err != nil {
		return nil, err
	}
	if a.Status == toStatus {
		return a, nil
	}
	if !model.ActivityCanTransition(a.Status, toStatus) {
		return nil, model.NewValidationError("status", "活动状态流转不合法")
	}
	a.Status = toStatus
	a.UpdatedAt = time.Now()
	if err := s.store.UpdateActivity(a); err != nil {
		return nil, err
	}
	return a, nil
}
