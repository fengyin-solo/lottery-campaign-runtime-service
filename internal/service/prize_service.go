package service

import (
	"sort"
	"strings"
	"time"

	"lottery/internal/model"
	"lottery/pkg/idgen"
)

func (s *Service) CreatePrize(input model.Prize) (*model.Prize, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetActivity(input.ActivityID); err != nil {
		return nil, model.NewValidationError("activity_id", "关联活动不存在")
	}
	p := &model.Prize{
		ID:         idgen.Hex(),
		ActivityID: input.ActivityID,
		Name:       input.Name,
		Total:      input.Total,
		Remaining:  input.Total,
		Weight:     input.Weight,
		Level:      input.Level,
		CreatedAt:  time.Now(),
	}
	if err := s.store.CreatePrize(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetPrize(id string) (*model.Prize, error) {
	return s.store.GetPrize(id)
}

func (s *Service) ListPrizes(filter model.PrizeFilter, page, size int) ([]*model.Prize, int, error) {
	all := s.store.ListPrizes()
	matched := make([]*model.Prize, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Level < matched[j].Level
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Prize{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdatePrize(id string, input model.Prize) (*model.Prize, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, model.NewValidationError("name", "奖品名称不能为空")
	}
	p, err := s.store.GetPrize(id)
	if err != nil {
		return nil, err
	}
	p.Name = input.Name
	p.Weight = input.Weight
	p.Level = input.Level
	if input.Total >= 0 {
		diff := input.Total - p.Total
		p.Total = input.Total
		p.Remaining += diff
		if p.Remaining < 0 {
			p.Remaining = 0
		}
	}
	if err := s.store.UpdatePrize(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePrize(id string) error {
	return s.store.DeletePrize(id)
}
