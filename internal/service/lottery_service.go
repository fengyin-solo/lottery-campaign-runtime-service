package service

import (
	"math/rand"
	"time"

	"lottery/internal/model"
	"lottery/pkg/idgen"
)

type DrawResult struct {
	Entry  *model.Entry  `json:"entry"`
	Winner *model.Winner `json:"winner,omitempty"`
	Prize  *model.Prize  `json:"prize,omitempty"`
}

func (s *Service) Draw(activityID, userID string) (*DrawResult, error) {
	if activityID == "" {
		return nil, model.NewValidationError("activity_id", "活动 ID 不能为空")
	}
	if userID == "" {
		return nil, model.NewValidationError("user_id", "用户 ID 不能为空")
	}

	act, err := s.store.GetActivity(activityID)
	if err != nil {
		return nil, err
	}
	if act.Status != model.ActivityActive {
		return nil, model.NewValidationError("status", "活动未在进行中")
	}

	entries, _, err := s.ListEntries(model.EntryFilter{ActivityID: activityID, UserID: userID}, 1, 10000)
	if err != nil {
		return nil, err
	}
	if act.LimitPerUser > 0 && len(entries) >= act.LimitPerUser {
		return nil, model.NewValidationError("limit", "已超过每人限抽次数")
	}

	prizes, _, err := s.ListPrizes(model.PrizeFilter{ActivityID: activityID}, 1, 10000)
	if err != nil {
		return nil, err
	}

	var available []*model.Prize
	for _, p := range prizes {
		if p.Remaining > 0 {
			available = append(available, p)
		}
	}

	var wonPrize *model.Prize
	if len(available) > 0 {
		totalWeight := 0
		for _, p := range available {
			totalWeight += p.Weight
		}
		if totalWeight > 0 {
			r := rand.Intn(totalWeight)
			for _, p := range available {
				r -= p.Weight
				if r < 0 {
					wonPrize = p
					break
				}
			}
		}
	}

	now := time.Now()
	entry := &model.Entry{
		ID:         idgen.Hex(),
		ActivityID: activityID,
		UserID:     userID,
		Result:     model.EntryResultLose,
		CreatedAt:  now,
	}

	var winner *model.Winner
	if wonPrize != nil {
		if err := s.store.DecrementPrizeRemaining(wonPrize.ID); err != nil {
			wonPrize = nil
		} else {
			entry.Result = model.EntryResultWin
			winner = &model.Winner{
				ID:         idgen.Hex(),
				ActivityID: activityID,
				PrizeID:    wonPrize.ID,
				UserID:     userID,
				Status:     model.WinnerStatusWon,
				WonAt:      now,
				CreatedAt:  now,
			}
			if err := s.store.CreateWinner(winner); err != nil {
				return nil, err
			}
		}
	}

	if err := s.store.CreateEntry(entry); err != nil {
		return nil, err
	}

	if u, err := s.store.GetUser(userID); err == nil {
		u.EntryCount++
		u.UpdatedAt = now
		_ = s.store.UpdateUser(u)
	}

	return &DrawResult{Entry: entry, Winner: winner, Prize: wonPrize}, nil
}

func (s *Service) BatchDraw(activityID, userID string, count int) ([]*DrawResult, error) {
	if count <= 0 {
		return nil, model.NewValidationError("count", "抽奖次数必须大于 0")
	}
	if count > 100 {
		return nil, model.NewValidationError("count", "单次批量抽奖不能超过 100 次")
	}
	results := make([]*DrawResult, 0, count)
	for i := 0; i < count; i++ {
		res, err := s.Draw(activityID, userID)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}
