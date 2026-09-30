package handlers

import (
	"net/http"

	"designreview/internal/config"
	"designreview/internal/middleware"
	"designreview/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler holds shared dependencies. DB sekarang *gorm.DB:
// query aplikasi lewat GORM, skema tetap dari migrations/*.sql.
type Handler struct {
	DB    *gorm.DB
	Store *storage.Store
	Cfg   config.Config
}

func New(db *gorm.DB, store *storage.Store, cfg config.Config) *Handler {
	return &Handler{DB: db, Store: store, Cfg: cfg}
}

func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": gin.H{"code": "not_implemented", "message": "scaffold only — not implemented yet"},
	})
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	v1 := r.Group("/api/v1")
	{
		// Designer auth (public).
		v1.POST("/auth/register", h.Register)
		v1.POST("/auth/login", h.Login)

		// Everything below requires Designer JWT (see docs/api.md).
		auth := v1.Group("")
		auth.Use(middleware.RequireDesigner(h.Cfg.JWTSecret))
		{
			auth.GET("/me", h.Me)

			auth.GET("/projects", notImplemented)
			auth.POST("/projects", notImplemented)
			auth.GET("/projects/:projectId", notImplemented)
			auth.PATCH("/projects/:projectId", notImplemented)
			auth.DELETE("/projects/:projectId", notImplemented)

			auth.POST("/projects/:projectId/share", notImplemented)
			auth.POST("/projects/:projectId/share/rotate", notImplemented)
			auth.DELETE("/projects/:projectId/share", notImplemented)

			auth.GET("/projects/:projectId/designs", notImplemented)
			auth.POST("/projects/:projectId/designs", notImplemented)

			auth.GET("/designs/:designId", notImplemented)
			auth.PATCH("/designs/:designId", notImplemented)
			auth.DELETE("/designs/:designId", notImplemented)

			auth.GET("/designs/:designId/versions", notImplemented)
			auth.POST("/designs/:designId/versions", notImplemented)

			auth.GET("/versions/:versionId", notImplemented)

			auth.GET("/versions/:versionId/pins", notImplemented)
			auth.POST("/versions/:versionId/pins", notImplemented)

			auth.GET("/pins/:pinId", notImplemented)
			auth.DELETE("/pins/:pinId", notImplemented)
			auth.POST("/pins/:pinId/comments", notImplemented)
			auth.PATCH("/pins/:pinId/status", notImplemented)

			auth.PATCH("/comments/:commentId", notImplemented)
			auth.DELETE("/comments/:commentId", notImplemented)
		}

		// Client surface: anonymous via share token (no JWT).
		// Invalid/revoked token -> 404 (indistinguishable), per docs/api.md.
		shared := v1.Group("/shared/:shareToken")
		{
			shared.GET("/project", notImplemented)
			shared.GET("/designs/:designId", notImplemented)
			shared.GET("/versions/:versionId", notImplemented)
			shared.GET("/pins/:pinId", notImplemented)
			shared.POST("/versions/:versionId/pins", notImplemented)
			shared.POST("/pins/:pinId/comments", notImplemented)
			shared.PATCH("/comments/:commentId", notImplemented)
			shared.DELETE("/comments/:commentId", notImplemented)
		}
	}
}
