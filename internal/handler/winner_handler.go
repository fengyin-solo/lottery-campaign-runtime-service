package handler

import (
	"net/http"

	"lottery/internal/model"
	"lottery/pkg/httpx"
)

func (s *Server) registerWinnerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/winners", s.createWinner)
	mux.HandleFunc("GET /api/winners", s.listWinners)
	mux.HandleFunc("GET /api/winners/{id}", s.getWinner)
	mux.HandleFunc("PUT /api/winners/{id}", s.updateWinner)
	mux.HandleFunc("DELETE /api/winners/{id}", s.deleteWinner)
	mux.HandleFunc("POST /api/winners/{id}/claim", s.claimWinner)
	mux.HandleFunc("POST /api/winners/batch-expire", s.batchExpireWinners)
}

type createWinnerRequest struct {
	ActivityID string `json:"activity_id"`
	PrizeID    string `json:"prize_id"`
	UserID     string `json:"user_id"`
	Status     string `json:"status"`
}

func (s *Server) createWinner(w http.ResponseWriter, r *http.Request) {
	var req createWinnerRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	winner, err := s.svc.CreateWinner(model.Winner{ActivityID: req.ActivityID, PrizeID: req.PrizeID, UserID: req.UserID, Status: req.Status})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, winner)
}

func (s *Server) listWinners(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.WinnerFilter{
		ActivityID: r.URL.Query().Get("activity_id"),
		UserID:     r.URL.Query().Get("user_id"),
		Status:     r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListWinners(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getWinner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	winner, err := s.svc.GetWinner(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, winner)
}

func (s *Server) updateWinner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req createWinnerRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	winner, err := s.svc.UpdateWinner(id, model.Winner{ActivityID: req.ActivityID, PrizeID: req.PrizeID, UserID: req.UserID, Status: req.Status})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, winner)
}

func (s *Server) deleteWinner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.svc.DeleteWinner(id); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) claimWinner(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	winner, err := s.svc.ClaimWinner(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, winner)
}

func (s *Server) batchExpireWinners(w http.ResponseWriter, r *http.Request) {
	count, err := s.svc.BatchExpireWinners()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]interface{}{"expired_count": count})
}
