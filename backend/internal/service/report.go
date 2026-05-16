package service

import (
	"context"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/repository"
)

type ReportService struct {
	couchRepo *repository.CouchDB
	cfg       *config.Config
}

func NewReportService(couchRepo *repository.CouchDB, cfg *config.Config) *ReportService {
	return &ReportService{
		couchRepo: couchRepo,
		cfg:       cfg,
	}
}

func (s *ReportService) RequestReport(ctx context.Context, sub, periodStart, periodEnd string) (string, error) {
	return "", nil
}

func (s *ReportService) GetJobs(ctx context.Context, sub string) ([]repository.ReportJobDoc, error) {
	return nil, nil
}

func (s *ReportService) GetJob(ctx context.Context, id, sub string) (*repository.ReportJobDoc, error) {
	return nil, nil
}

func (s *ReportService) GetPDF(ctx context.Context, id, sub string) ([]byte, error) {
	return nil, nil
}
