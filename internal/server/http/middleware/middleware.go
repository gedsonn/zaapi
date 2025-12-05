package middleware

import (
	"github.com/gedsonn/zaapi/internal/maneger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AttachManager(m *maneger.Manager) gin.HandlerFunc {
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

func ExtractManeger(c *gin.Context) *maneger.Manager {
	v, ok := c.Get("manager")
	if !ok {
		panic("ok")
	}
	return v.(*maneger.Manager)
}

