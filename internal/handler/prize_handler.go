package handler

import (
	"net/http"

	"lottery/internal/model"
	"lottery/pkg/httpx"
)

func (s *Server) registerPrizeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/prizes", s.createPrize)
	mux.HandleFunc("GET /api/prizes", s.listPrizes)
	mux.HandleFunc("GET /api/prizes/{id}", s.getPrize)
	mux.HandleFunc("PUT /api/prizes/{id}", s.updatePrize)
	mux.HandleFunc("DELETE /api/prizes/{id}", s.deletePrize)
}

type createPrizeRequest struct {
	ActivityID string `json:"activity_id"`
	Name       string `json:"name"`
	Total      int    `json:"total"`
	Weight     int    `json:"weight"`
	Level      int    `json:"level"`
}

func (s *Server) createPrize(w http.ResponseWriter, r *http.Request) {
	var req createPrizeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.CreatePrize(model.Prize{ActivityID: req.ActivityID, Name: req.Name, Total: req.Total, Weight: req.Weight, Level: req.Level})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

func (s *Server) listPrizes(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.PrizeFilter{
		ActivityID: r.URL.Query().Get("activity_id"),
		Keyword:    r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListPrizes(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getPrize(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.svc.GetPrize(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) updatePrize(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req createPrizeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UpdatePrize(id, model.Prize{Name: req.Name, Total: req.Total, Weight: req.Weight, Level: req.Level})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) deletePrize(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.svc.DeletePrize(id); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
