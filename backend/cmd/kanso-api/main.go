package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"

	"github.com/edson/kanso-api/internal/config"
	"github.com/edson/kanso-api/internal/handler"
	"github.com/edson/kanso-api/internal/middleware"
	"github.com/edson/kanso-api/internal/nlp"
	"github.com/edson/kanso-api/internal/pdf"
	"github.com/edson/kanso-api/internal/repository"
	"github.com/edson/kanso-api/internal/service"
)

func ensureCouchDBDatabases(cfg *config.Config) error {
	dbs := []string{repository.DBRegistros, repository.DBSentimentos, repository.DBPreferencias, repository.DBRelatorios, repository.DBUsuarios, repository.DBTreinamento}
	for _, db := range dbs {
		url := cfg.CouchDBURL + "/" + db
		req, _ := http.NewRequest("PUT", url, bytes.NewReader([]byte{}))
		req.SetBasicAuth(cfg.CouchDBUser, cfg.CouchDBPass)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Warn("could not create database", "db", db, "error", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusCreated {
			slog.Info("database created", "db", db)
		}

		// Allow any authenticated user (via JWT) to sync — req for PouchDB live sync
		secURL := cfg.CouchDBURL + "/" + db + "/_security"
		secBody := `{"members":{"roles":[]},"admins":{"roles":["_admin"]}}`
		secReq, _ := http.NewRequest("PUT", secURL, bytes.NewReader([]byte(secBody)))
		secReq.SetBasicAuth(cfg.CouchDBUser, cfg.CouchDBPass)
		secReq.Header.Set("Content-Type", "application/json")
		secResp, secErr := http.DefaultClient.Do(secReq)
		if secErr != nil {
			slog.Warn("could not set security", "db", db, "error", secErr)
		} else {
			secResp.Body.Close()
		}
	}
	return nil
}

func ensureCouchDBIndexes(cfg *config.Config) {
	indexes := []struct {
		db     string
		name   string
		fields []string
	}{
		{repository.DBRelatorios, "idx-type-user-createdat", []string{"type", "userSub", "createdAt"}},
		{repository.DBRegistros, "idx-type-user-datahora", []string{"type", "userId", "dataHora"}},
	}
	for _, idx := range indexes {
		body := `{"index":{"fields":["` +
			idx.fields[0] + `","` + idx.fields[1] + `","` + idx.fields[2] +
			`"]},"name":"` + idx.name + `","type":"json"}`
		url := cfg.CouchDBURL + "/" + idx.db + "/_index"
		req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))
		req.SetBasicAuth(cfg.CouchDBUser, cfg.CouchDBPass)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Warn("could not create index", "name", idx.name, "db", idx.db, "error", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			slog.Info("index created", "name", idx.name, "db", idx.db)
		} else {
			slog.Warn("index creation returned unexpected status", "name", idx.name, "db", idx.db, "status", resp.StatusCode)
		}
	}
}

func ensureValidateDocUpdate(cfg *config.Config) {
	validateFn := `function(newDoc, oldDoc, userCtx, secObj) {
    if (userCtx.roles.indexOf('_admin') !== -1) { return true; }
    if (newDoc.type === 'config') { return true; }
    if (newDoc.userSub && userCtx.name && newDoc.userSub === userCtx.name) { return true; }
    if (newDoc.userId && userCtx.name && newDoc.userId === userCtx.name) { return true; }
    if (userCtx.name && newDoc._id === 'user:' + userCtx.name) { return true; }
    throw({forbidden: 'document access denied'});
}`
	ddoc := map[string]interface{}{
		"_id":                "_design/security",
		"language":           "javascript",
		"validate_doc_update": validateFn,
	}
	body, _ := json.Marshal(ddoc)

	dbs := []string{repository.DBRegistros, repository.DBSentimentos, repository.DBPreferencias, repository.DBRelatorios, repository.DBUsuarios, repository.DBTreinamento}
	for _, db := range dbs {
		url := cfg.CouchDBURL + "/" + db + "/_design/security"
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(body))
		req.SetBasicAuth(cfg.CouchDBUser, cfg.CouchDBPass)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Warn("could not set validate_doc_update", "db", db, "error", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			slog.Info("validate_doc_update deployed", "db", db)
		} else {
			slog.Warn("validate_doc_update returned unexpected status", "db", db, "status", resp.StatusCode)
		}
	}
}

func pushAuthMiddleware(jwtSecret []byte, apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key != "" && key == apiKey {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				slog.Warn("push: missing authorization",
					"path", r.URL.Path,
					"ip", r.RemoteAddr,
				)
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				slog.Warn("push: invalid JWT",
					"path", r.URL.Path,
					"ip", r.RemoteAddr,
					"error", err,
				)
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			claims := token.Claims.(jwt.MapClaims)
			role, _ := claims["role"].(string)
			if role != "admin" {
				slog.Warn("push: not admin role",
					"path", r.URL.Path,
					"ip", r.RemoteAddr,
				)
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), middleware.UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	cfg := config.Load()

	couchRepo := repository.NewCouchDB(cfg.CouchDBURL, cfg.CouchDBUser, cfg.CouchDBPass)
	authSvc := service.NewAuth(cfg.GoogleClienID, cfg.JWTSecret, couchRepo)
	authHandler := handler.NewAuth(authSvc, cfg.JWTSecret)

	chromedpURL := os.Getenv("CHROMEDP_WS_URL")
	var pdfGen *pdf.Generator
	if chromedpURL != "" {
		pdfGen = pdf.NewRemoteGenerator(chromedpURL, 30*time.Second)
		slog.Info("pdf: using remote chromedp", "url", chromedpURL)
	} else {
		pdfGen = pdf.NewGenerator("", 30*time.Second)
		slog.Info("pdf: using local chromedp")
	}
	reportSvc := service.NewReportService(couchRepo, pdfGen, cfg)
	reportHandler := handler.NewReportHandler(reportSvc, cfg)
	pushSvc := service.NewPushService(couchRepo, cfg.FCMServerKey, cfg.FCMProjectID, cfg.FCMServiceAccountB64)
	pushHandler := handler.NewPushHandler(pushSvc)

	// Start NLP watcher if NLP service is available
	grpcCACert := os.Getenv("GRPC_CA_CERT")
	nlpClient, err := nlp.NewClient(cfg.NLPGrpAddr, grpcCACert)
	if err != nil {
		slog.Warn("nlp client not available — watcher not started", "error", err)
	} else {
		watcherSvc := service.NewWatcherService(couchRepo, nlpClient, cfg)
		watcherSvc.Start()
		slog.Info("NLP watcher started", "addr", cfg.NLPGrpAddr)
	}

	// Create treinamento service (works even without NLP client)
	treinamentoSvc := service.NewTreinamentoService(couchRepo, nlpClient, cfg)
	treinamentoHandler := handler.NewTreinamentoHandler(treinamentoSvc)

	// Start weekly training scheduler
	trainInterval := 168 * time.Hour // default: 7 days
	if v := os.Getenv("TRAIN_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			trainInterval = d
		}
	}
	treinamentoSvc.StartScheduler(context.Background(), trainInterval)

	ensureCouchDBDatabases(cfg)
	ensureCouchDBIndexes(cfg)
	ensureValidateDocUpdate(cfg)

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
		r.Post("/api/train", treinamentoHandler.HandleTrain)
		r.Get("/api/train/status", treinamentoHandler.HandleTrainStatus)
		r.Post("/api/reanalyze", treinamentoHandler.HandleReanalyze)
	})

	// Push send — JWT admin role OR scheduler API key
	r.With(pushAuthMiddleware([]byte(cfg.JWTSecret), cfg.SchedulerAPIKey)).Post("/api/push/send", pushHandler.HandleSend)

	slog.Info("starting server", "port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
