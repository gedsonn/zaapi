package middleware

import (
	manager "github.com/gedsonn/zaapi/maneger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AttachManager(m *manager.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("manager", m)
		c.Next()
	}
}

func AttachRequestId() gin.HandlerFunc {
	id := uuid.New().String()

	return func(ctx *gin.Context) {
		ctx.Set("request_id", id)
		ctx.Writer.Header().Set("X-Request-ID", id)
		ctx.Next()
	}

}

func ExtractManeger(c *gin.Context) *manager.SessionManager {
	v, ok := c.Get("manager")
	if !ok {
		panic("ok")
	}
	return v.(*manager.SessionManager)
}

