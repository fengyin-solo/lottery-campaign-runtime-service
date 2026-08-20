package handler

import (
	"net/http"

	"lottery/internal/model"
	"lottery/pkg/httpx"
)

func (s *Server) registerActivityRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/activities", s.createActivity)
	mux.HandleFunc("GET /api/activities", s.listActivities)
	mux.HandleFunc("GET /api/activities/{id}", s.getActivity)
	mux.HandleFunc("PUT /api/activities/{id}", s.updateActivity)
	mux.HandleFunc("DELETE /api/activities/{id}", s.deleteActivity)
	mux.HandleFunc("POST /api/activities/{id}/transition", s.transitionActivity)
}

type createActivityRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Status       string `json:"status"`
	LimitPerUser int    `json:"limit_per_user"`
}

func (s *Server) createActivity(w http.ResponseWriter, r *http.Request) {
	var req createActivityRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateActivity(model.Activity{Name: req.Name, Description: req.Description, Status: req.Status, LimitPerUser: req.LimitPerUser})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listActivities(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ActivityFilter{
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListActivities(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a, err := s.svc.GetActivity(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) updateActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req createActivityRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateActivity(id, model.Activity{Name: req.Name, Description: req.Description, Status: req.Status, LimitPerUser: req.LimitPerUser})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.svc.DeleteActivity(id); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionRequest struct {
	Status string `json:"status"`
}

func (s *Server) transitionActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.TransitionActivityStatus(id, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}
