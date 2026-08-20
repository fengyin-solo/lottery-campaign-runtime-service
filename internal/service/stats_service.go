package service

import (
	"sort"

	"lottery/internal/model"
)

type ActivityStats struct {
	ActivityID     string            `json:"activity_id"`
	EntryCount     int               `json:"entry_count"`
	UniqueUsers    int               `json:"unique_users"`
	WinnerCount    int               `json:"winner_count"`
	PrizeDistribution map[string]int `json:"prize_distribution"`
}

type PrizeStats struct {
	PrizeID      string `json:"prize_id"`
	PrizeName    string `json:"prize_name"`
	TotalWins    int    `json:"total_wins"`
	ClaimedCount int    `json:"claimed_count"`
	ExpiredCount int    `json:"expired_count"`
}

func (s *Service) GetActivityStats(activityID string) (*ActivityStats, error) {
	if activityID == "" {
		return nil, nil
	}
	entries, _, err := s.ListEntries(model.EntryFilter{ActivityID: activityID}, 1, 100000)
	if err != nil {
		return nil, err
	}
	winners, _, err := s.ListWinners(model.WinnerFilter{ActivityID: activityID}, 1, 100000)
	if err != nil {
		return nil, err
	}

	userSet := make(map[string]struct{})
	for _, e := range entries {
		userSet[e.UserID] = struct{}{}
	}

	prizeDist := make(map[string]int)
	for _, w := range winners {
		prizeDist[w.PrizeID]++
	}

	return &ActivityStats{
		ActivityID:        activityID,
		EntryCount:        len(entries),
		UniqueUsers:       len(userSet),
		WinnerCount:       len(winners),
		PrizeDistribution: prizeDist,
	}, nil
}

func (s *Service) GetPrizeStats(activityID string) ([]*PrizeStats, error) {
	prizes, _, err := s.ListPrizes(model.PrizeFilter{ActivityID: activityID}, 1, 10000)
	if err != nil {
		return nil, err
	}
	winners, _, err := s.ListWinners(model.WinnerFilter{ActivityID: activityID}, 1, 100000)
	if err != nil {
		return nil, err
	}

	prizeMap := make(map[string]*model.Prize)
	for _, p := range prizes {
		prizeMap[p.ID] = p
	}

	statsMap := make(map[string]*PrizeStats)
	for _, p := range prizes {
		statsMap[p.ID] = &PrizeStats{
			PrizeID:   p.ID,
			PrizeName: p.Name,
		}
	}

	for _, w := range winners {
		st, ok := statsMap[w.PrizeID]
		if !ok {
			name := w.PrizeID
			if p, ok2 := prizeMap[w.PrizeID]; ok2 {
				name = p.Name
			}
			st = &PrizeStats{PrizeID: w.PrizeID, PrizeName: name}
			statsMap[w.PrizeID] = st
		}
		st.TotalWins++
		if w.Status == "claimed" {
			st.ClaimedCount++
		} else if w.Status == "expired" {
			st.ExpiredCount++
		}
	}

	result := make([]*PrizeStats, 0, len(statsMap))
	for _, st := range statsMap {
		result = append(result, st)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PrizeID < result[j].PrizeID
	})
	return result, nil
}
