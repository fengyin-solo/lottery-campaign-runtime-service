package handler

import (
	"net/http"

	"lottery/pkg/httpx"
)

func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/lottery/draw", s.draw)
	mux.HandleFunc("POST /api/lottery/batch-draw", s.batchDraw)
	mux.HandleFunc("GET /api/stats/activities/{id}", s.getActivityStats)
	mux.HandleFunc("GET /api/stats/prizes", s.getPrizeStats)
}

type drawRequest struct {
	ActivityID string `json:"activity_id"`
	UserID     string `json:"user_id"`
}

type batchDrawRequest struct {
	ActivityID string `json:"activity_id"`
	UserID     string `json:"user_id"`
	Count      int    `json:"count"`
}

func (s *Server) draw(w http.ResponseWriter, r *http.Request) {
	var req drawRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	res, err := s.svc.Draw(req.ActivityID, req.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, res)
}

func (s *Server) batchDraw(w http.ResponseWriter, r *http.Request) {
	var req batchDrawRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	results, err := s.svc.BatchDraw(req.ActivityID, req.UserID, req.Count)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, results)
}

func (s *Server) getActivityStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stats, err := s.svc.GetActivityStats(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) getPrizeStats(w http.ResponseWriter, r *http.Request) {
	activityID := r.URL.Query().Get("activity_id")
	stats, err := s.svc.GetPrizeStats(activityID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}
