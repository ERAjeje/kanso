package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/handler"
	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/nlp"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/service"
)

func ensureCouchDBDatabases(cfg *config.Config) error {
	dbs := []string{"registros", "sentimentos", "preferencias"}
	for _, db := range dbs {
		url := cfg.CouchDBURL + "/" + db
		req, _ := http.NewRequest("PUT", url, bytes.NewReader([]byte{}))
		req.SetBasicAuth(cfg.CouchDBUser, cfg.CouchDBPass)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("warning: could not create database %s: %v", db, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusCreated {
			log.Printf("database %s created", db)
		}
	}
	return nil
}

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

	// Start NLP watcher if NLP service is available
	nlpClient, err := nlp.NewClient(cfg.NLPGrpAddr)
	if err != nil {
		log.Printf("warning: nlp client not available — watcher not started: %v", err)
	} else {
		watcherSvc := service.NewWatcherService(couchRepo, nlpClient, cfg)
		watcherSvc.Start()
	}

	ensureCouchDBDatabases(cfg)

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
	})

	// Internal routes (scheduler access only — on Docker internal network)
	r.Post("/api/push/send", pushHandler.HandleSend)

	log.Printf("Starting server on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
