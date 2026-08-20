package service

import (
	"sort"
	"time"

	"lottery/internal/model"
	"lottery/pkg/idgen"
)

func (s *Service) CreateWinner(input model.Winner) (*model.Winner, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetActivity(input.ActivityID); err != nil {
		return nil, model.NewValidationError("activity_id", "关联活动不存在")
	}
	if _, err := s.store.GetPrize(input.PrizeID); err != nil {
		return nil, model.NewValidationError("prize_id", "关联奖品不存在")
	}
	w := &model.Winner{
		ID:         idgen.Hex(),
		ActivityID: input.ActivityID,
		PrizeID:    input.PrizeID,
		UserID:     input.UserID,
		Status:     model.WinnerStatusWon,
		WonAt:      time.Now(),
		CreatedAt:  time.Now(),
	}
	if err := s.store.CreateWinner(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Service) GetWinner(id string) (*model.Winner, error) {
	return s.store.GetWinner(id)
}

func (s *Service) ListWinners(filter model.WinnerFilter, page, size int) ([]*model.Winner, int, error) {
	all := s.store.ListWinners()
	matched := make([]*model.Winner, 0, len(all))
	for _, w := range all {
		if filter.Match(w) {
			matched = append(matched, w)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].WonAt.After(matched[j].WonAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Winner{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateWinner(id string, input model.Winner) (*model.Winner, error) {
	w, err := s.store.GetWinner(id)
	if err != nil {
		return nil, err
	}
	if input.Status != "" && input.Status != w.Status {
		if !model.WinnerCanTransition(w.Status, input.Status) {
			return nil, model.NewValidationError("status", "中奖状态流转不合法")
		}
		w.Status = input.Status
		if w.Status == model.WinnerStatusClaimed {
			w.ClaimedAt = time.Now()
		}
	}
	if err := s.store.UpdateWinner(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Service) DeleteWinner(id string) error {
	return s.store.DeleteWinner(id)
}

func (s *Service) ClaimWinner(id string) (*model.Winner, error) {
	w, err := s.store.GetWinner(id)
	if err != nil {
		return nil, err
	}
	if w.Status != model.WinnerStatusWon {
		return nil, model.NewValidationError("status", "只能领取未领取的中奖记录")
	}
	w.Status = model.WinnerStatusClaimed
	w.ClaimedAt = time.Now()
	if err := s.store.UpdateWinner(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Service) BatchExpireWinners() (int, error) {
	all := s.store.ListWinners()
	count := 0
	for _, w := range all {
		if w.Status == model.WinnerStatusWon {
			w.Status = model.WinnerStatusExpired
			if err := s.store.UpdateWinner(w); err == nil {
				count++
			}
		}
	}
	return count, nil
}
