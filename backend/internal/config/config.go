package config

import "os"

type Config struct {
	Port                 string
	CouchDBURL           string
	CouchDBUser          string
	CouchDBPass          string
	JWTSecret            string
	GoogleClienID        string
	PDFTmpDir            string
	FCMServerKey         string
	FCMProjectID         string
	FCMServiceAccountB64 string
	NLPGrpAddr           string
	NLPHTTPAddr          string
	SchedulerAPIKey      string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		CouchDBURL:    getEnv("COUCHDB_URL", "http://couchdb:5984"),
		CouchDBUser:   getEnv("COUCHDB_USER", "admin"),
		CouchDBPass:   getEnv("COUCHDB_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		GoogleClienID: getEnv("GOOGLE_CLIENT_ID", ""),
		PDFTmpDir:     getEnv("PDF_TMP_DIR", "/tmp/kanso-pdf"),
		FCMServerKey:         getEnv("FCM_SERVER_KEY", ""),
		FCMProjectID:         getEnv("FCM_PROJECT_ID", ""),
		FCMServiceAccountB64: getEnv("FCM_SERVICE_ACCOUNT_B64", ""),
		NLPGrpAddr:           getEnv("NLP_GRPC_ADDR", "nlp:50051"),
		NLPHTTPAddr:          getEnv("NLP_HTTP_ADDR", "http://nlp:8000"),
		SchedulerAPIKey:      getEnv("SCHEDULER_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
