package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	FrontendURL  string
	MinioEndpoint string
	MinioUser    string
	MinioPass    string
	MinioBucket  string
	MinioUseSSL  bool
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		Port:         getenv("PORT", "8080"),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://review:reviewpw@localhost:5432/designreview?sslmode=disable"),
		JWTSecret:    getenv("JWT_SECRET", "change-me-in-production"),
		FrontendURL:  getenv("FRONTEND_URL", "http://localhost:5173"),
		MinioEndpoint: getenv("MINIO_ENDPOINT", "localhost:9000"),
		MinioUser:    getenv("MINIO_ROOT_USER", "minioadmin"),
		MinioPass:    getenv("MINIO_ROOT_PASSWORD", "minioadminpw"),
		MinioBucket:  getenv("MINIO_BUCKET", "designs"),
		MinioUseSSL:  getenv("MINIO_USE_SSL", "false") == "true",
	}
}
