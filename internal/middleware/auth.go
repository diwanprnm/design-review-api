package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// DesignerIDKey is the gin context key for the authenticated designer id.
const DesignerIDKey = "designer_id"

// RequireDesigner enforces `Authorization: Bearer <jwt>` on designer routes.
// Scaffold phase: signature + behavior contract only; full claims validation
// lands with the auth implementation (see handlers/auth.go).
func RequireDesigner(secret string) gin.HandlerFunc {
	_ = secret
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "unauthorized", "message": "missing bearer token"},
			})
			return
		}
		// TODO(auth): parse + verify JWT with secret, load designer, 401 on failure.
		var _ = jwt.ErrTokenExpired
		c.Set(DesignerIDKey, "todo-designer-id")
		c.Next()
	}
}
