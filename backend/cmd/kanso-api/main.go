package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/handler"
	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/service"
)

func main() {
	cfg := config.Load()

	couchRepo := repository.NewCouchDB(cfg.CouchDBURL, cfg.CouchDBUser, cfg.CouchDBPass)
	authSvc := service.NewAuth(cfg.GoogleClienID, cfg.JWTSecret, couchRepo)
	authHandler := handler.NewAuth(authSvc, cfg.JWTSecret)

	pdfGen := pdf.NewGenerator("", 30*time.Second)
	reportSvc := service.NewReportService(couchRepo, pdfGen, cfg)
	reportHandler := handler.NewReportHandler(reportSvc, cfg)
	pushSvc := service.NewPushService(couchRepo, cfg.FCMServerKey)
	pushHandler := handler.NewPushHandler(pushSvc)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://kanso.local"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.Timeout(30 * time.Second))

	r.Get("/api/health", handler.Health)

	r.Post("/api/auth/google", authHandler.HandleGoogleLogin)
	r.Post("/api/auth/refresh", authHandler.HandleRefresh)
	r.Post("/api/auth/logout", authHandler.HandleLogout)

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTRequired([]byte(cfg.JWTSecret)))
		r.Get("/api/auth/me", authHandler.HandleMe)
		r.Post("/api/reports", reportHandler.HandleRequestReport)
		r.Get("/api/reports", reportHandler.HandleListReports)
		r.Get("/api/reports/{id}", reportHandler.HandleGetReport)
		r.Get("/api/reports/{id}/download", reportHandler.HandleDownload)
		r.Post("/api/push/subscribe", pushHandler.HandleSubscribe)
		r.Get("/api/push/preferences", pushHandler.HandleGetPreferences)
		r.Put("/api/push/preferences", pushHandler.HandleUpdatePreferences)
	})

	// Internal routes (scheduler access only — on Docker internal network)
	r.Post("/api/push/send", pushHandler.HandleSend)

	log.Printf("Starting server on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
