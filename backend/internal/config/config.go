package config

import "os"

type Config struct {
	Port          string
	CouchDBURL    string
	CouchDBUser   string
	CouchDBPass   string
	JWTSecret     string
	GoogleClienID string
	PDFTmpDir     string
	FCMServerKey  string
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
		FCMServerKey:  getEnv("FCM_SERVER_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
