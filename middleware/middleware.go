package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"issue-pm/config"
	"issue-pm/pkg/auth"
	"issue-pm/pkg/response"
	"log"
	"net/http"
	"strings"
	"time"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
func BodyLimit(bytes int64) gin.HandlerFunc {
	return func(c *gin.Context) { c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, bytes); c.Next() }
}
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("request_id=%s method=%s path=%s status=%d duration=%s", c.GetString("request_id"), c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				response.Error(c, 500, response.CodeInternal, "internal server error")
			}
		}()
		c.Next()
	}
}
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
func JWT(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			response.Error(c, 401, response.CodeUnauthorized, "missing bearer token")
			return
		}
		cl, err := auth.Parse(strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")), cfg.JWTSecret)
		if err != nil {
			response.Error(c, 401, response.CodeUnauthorized, "invalid or expired token")
			return
		}
		c.Set("user_id", cl.UserID)
		c.Set("role", cl.Role)
		c.Next()
	}
}
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != "admin" {
			response.Error(c, 403, response.CodeForbidden, "admin role required")
			return
		}
		c.Next()
	}
}
