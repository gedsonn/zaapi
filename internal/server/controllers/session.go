package controllers

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gedsonn/zaapi/internal/maneger"
	"github.com/gedsonn/zaapi/internal/server/http/middleware"
	"github.com/gin-gonic/gin"
)

func CreateSession(ctx *gin.Context) {
	node, _ := snowflake.NewNode(1)
	m := middleware.ExtractManeger(ctx)

	id := node.Generate()

	i, err := maneger.CreateInstance(id.String())
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	m.Add(i)
	i.Start()

	ctx.JSON(200, gin.H{
		"message":    "Sessão criada com sucesso",
		"session_id": i.Id,
	})
}

func SessionQRcode(ctx *gin.Context) {
	m := middleware.ExtractManeger(ctx)
	id := ctx.Param("session")

	instance, ok := m.Get(id)
	if !ok {
		ctx.JSON(404, gin.H{
			"error": "Instancia não encontrada",
		})
		return
	}

	if instance.Client.IsLoggedIn() {
		ctx.JSON(409, gin.H{"error": "Esta sessão já está conectada."})
		return
	}

	qr, err := instance.GetQR()
	if err != nil {
		if err.Error() == "solicitando novo QR code, tente novamente em alguns segundos" {
			ctx.JSON(200, gin.H{
				"message": err.Error(),
				"code":    "ZAAPI-0001", //0001 abrir "canal"
			})
			return
		}
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	if qr == nil {
		qr, err = instance.GetQR()
		if err != nil {
			if err.Error() == "solicitando novo QR code, tente novamente em alguns segundos" {
				ctx.JSON(200, gin.H{
					"message": err.Error(),
					"code":    "ZAAPI-0001", //0001 abrir "canal"
				})
				return
			}
			ctx.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// O QR code do WhatsApp geralmente expira em menos de 1 minuto.
	// Um valor seguro é em torno de 45 segundos.
	expiresIn := 45 - int(time.Since(instance.LastQRTime).Seconds())
	ctx.JSON(200, gin.H{
		"base64":     qr.Base64,
		"expires_in": expiresIn,
	})
}
