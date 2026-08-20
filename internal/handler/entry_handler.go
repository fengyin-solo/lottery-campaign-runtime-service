package handler

import (
	"net/http"

	"lottery/internal/model"
	"lottery/pkg/httpx"
)

func (s *Server) registerEntryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/entries", s.createEntry)
	mux.HandleFunc("GET /api/entries", s.listEntries)
	mux.HandleFunc("GET /api/entries/{id}", s.getEntry)
	mux.HandleFunc("PUT /api/entries/{id}", s.updateEntry)
	mux.HandleFunc("DELETE /api/entries/{id}", s.deleteEntry)
}

type createEntryRequest struct {
	ActivityID string `json:"activity_id"`
	UserID     string `json:"user_id"`
	Result     string `json:"result"`
}

func (s *Server) createEntry(w http.ResponseWriter, r *http.Request) {
	var req createEntryRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.CreateEntry(model.Entry{ActivityID: req.ActivityID, UserID: req.UserID, Result: req.Result})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, e)
}

func (s *Server) listEntries(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.EntryFilter{
		ActivityID: r.URL.Query().Get("activity_id"),
		UserID:     r.URL.Query().Get("user_id"),
		Result:     r.URL.Query().Get("result"),
	}
	items, total, err := s.svc.ListEntries(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	e, err := s.svc.GetEntry(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

func (s *Server) updateEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req createEntryRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	e, err := s.svc.UpdateEntry(id, model.Entry{ActivityID: req.ActivityID, UserID: req.UserID, Result: req.Result})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, e)
}

func (s *Server) deleteEntry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.svc.DeleteEntry(id); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
