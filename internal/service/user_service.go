package service

import (
	"sort"
	"time"

	"lottery/internal/model"
	"lottery/pkg/idgen"
)

func (s *Service) CreateUser(input model.User) (*model.User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	u := &model.User{
		ID:         idgen.Hex(),
		Name:       input.Name,
		EntryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.store.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) GetUser(id string) (*model.User, error) {
	return s.store.GetUser(id)
}

func (s *Service) ListUsers(filter model.UserFilter, page, size int) ([]*model.User, int, error) {
	all := s.store.ListUsers()
	matched := make([]*model.User, 0, len(all))
	for _, u := range all {
		if filter.Match(u) {
			matched = append(matched, u)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.User{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateUser(id string, input model.User) (*model.User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	u, err := s.store.GetUser(id)
	if err != nil {
		return nil, err
	}
	u.Name = input.Name
	u.EntryCount = input.EntryCount
	u.UpdatedAt = time.Now()
	if err := s.store.UpdateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) DeleteUser(id string) error {
	return s.store.DeleteUser(id)
}
