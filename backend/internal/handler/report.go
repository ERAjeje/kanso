package handler

import (
	"net/http"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/service"
)

type ReportHandler struct {
	svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService, cfg *config.Config) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) HandleRequestReport(w http.ResponseWriter, r *http.Request) {
}

func (h *ReportHandler) HandleListReports(w http.ResponseWriter, r *http.Request) {
}

func (h *ReportHandler) HandleGetReport(w http.ResponseWriter, r *http.Request) {
}

func (h *ReportHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
}
