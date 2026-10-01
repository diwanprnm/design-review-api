package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)


const DesignerIDKey = "designer_id"


func RequireDesigner(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		raw, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || raw == "" {
			unauthorized(c, "missing bearer token")
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
			// Tolak alg selain HS256 — mencegah serangan alg confusion.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			unauthorized(c, "invalid or expired token")
			return
		}

		// Claim di-encode sebagai string (UUID) oleh Login; JSON number tetap
		// ditangani agar token lama tidak langsung pecah.
		var designerID string
		switch v := claims["designer_id"].(type) {
		case string:
			designerID = v
		case float64:
			designerID = strconv.FormatInt(int64(v), 10)
		}
		if designerID == "" {
			unauthorized(c, "invalid token claims")
			return
		}

		c.Set(DesignerIDKey, designerID)
		c.Next()
	}
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{"code": "unauthorized", "message": message},
	})
}
