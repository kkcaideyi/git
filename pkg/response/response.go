package response

import "github.com/gin-gonic/gin"

const (
	CodeOK           = 0
	CodeInvalid      = 40001
	CodeUnauthorized = 40101
	CodeForbidden    = 40301
	CodeNotFound     = 40401
	CodeConflict     = 40901
	CodeInternal     = 50001
)

type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, Envelope{Code: CodeOK, Message: "ok", Data: data, RequestID: c.GetString("request_id")})
}
func Created(c *gin.Context, data any) {
	c.JSON(201, Envelope{Code: CodeOK, Message: "created", Data: data, RequestID: c.GetString("request_id")})
}
func Error(c *gin.Context, status, code int, message string) {
	c.AbortWithStatusJSON(status, Envelope{Code: code, Message: message, RequestID: c.GetString("request_id")})
}
