package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/service"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService, cfg *config.Config) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// HandleRequestReport handles POST /api/reports
// Creates a new report generation job and returns 202 with job ID.
// The backend computes periodStart (from last completed report) and periodEnd (current time).
func (h *ReportHandler) HandleRequestReport(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	jobID, err := h.svc.RequestReport(r.Context(), sub)
	if err != nil {
		log.Printf("failed to request report: %v", err)
		http.Error(w, `{"error":"failed to create report job"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobID})
}

// HandleListReports handles GET /api/reports
// Returns all report jobs for the authenticated user.
func (h *ReportHandler) HandleListReports(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)

	jobs, err := h.svc.GetJobs(r.Context(), sub)
	if err != nil {
		log.Printf("failed to list reports: %v", err)
		http.Error(w, `{"error":"failed to list reports"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// HandleGetReport handles GET /api/reports/{id}
// Returns a single report job with ownership check.
func (h *ReportHandler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)
	jobID := chi.URLParam(r, "id")

	job, err := h.svc.GetJob(r.Context(), jobID, sub)
	if err != nil {
		log.Printf("failed to get report: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// HandleDownload handles GET /api/reports/{id}/download
// Returns the PDF file for a completed report with ownership check.
func (h *ReportHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	sub, _ := claims["sub"].(string)
	jobID := chi.URLParam(r, "id")

	data, err := h.svc.GetPDF(r.Context(), jobID, sub)
	if err != nil {
		log.Printf("failed to get PDF: %v", err)
		http.Error(w, `{"error":"failed to get PDF"}`, http.StatusInternalServerError)
		return
	}
	if data == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+jobID+".pdf\"")
	w.Write(data)
}
