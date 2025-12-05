package http

import (
	"fmt"

	"github.com/apex/log"
	"github.com/gedsonn/zaapi/internal/maneger"
	"github.com/gedsonn/zaapi/internal/server/controllers"
	"github.com/gedsonn/zaapi/internal/server/http/middleware"
	"github.com/gin-gonic/gin"
)

func Configure(m *maneger.Manager) *gin.Engine {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.AttachRequestId())
	router.Use(middleware.AttachManager(m))

	router.Use(gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		requestID := ""
		if params.Keys != nil {
			if v, ok := params.Keys["request_id"]; ok {
				requestID = fmt.Sprintf("%v", v)
			}
		}

		log.WithFields(log.Fields{
			"request_id": requestID,
			"method":     params.Method,
			"path":       params.Path,
			"status":     params.StatusCode,
			"latency":    params.Latency,
		}).Debug("HTTP Request")
		return ""
	}))

	log.SetLevel(log.DebugLevel)


	router.POST("/", controllers.CreateSession)
	
	session := router.Group("/:session") 
	{
		session.GET("/qr", controllers.SessionQRcode)
	}
	

	return router
}
