package main

import (
	"log"

	"designreview/internal/config"
	"designreview/internal/db"
	"designreview/internal/handlers"
	"designreview/internal/storage"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg := config.Load()

	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	// GORM membungkus database/sql: ambil handle aslinya untuk Close().
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatalf("db handle: %v", err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(gdb, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	store, err := storage.NewMinio(cfg)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}

	r := gin.Default()

	// TODO(frontend): tighten for production; permissive for local Vue dev.
	r.Use(cors(cfg.FrontendURL))

	h := handlers.New(gdb, store, cfg)
	h.RegisterRoutes(r)

	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

func cors(frontendURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", frontendURL)
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Client-Token")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}
		c.Next()
	}
}
