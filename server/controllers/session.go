package controllers

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gedsonn/zaapi/server/http/middleware"
	"github.com/gin-gonic/gin"
)

func CreateSession(ctx *gin.Context) {
	node, _ := snowflake.NewNode(1)
	m := middleware.ExtractManeger(ctx)

	id := node.Generate()

	go func() {
		_, err := m.CreateSession(id.String())
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
	}()

	ctx.JSON(200, gin.H{
		"message":    "Sessão criada com sucesso",
		"session_id": id,
	})
}

func SessionQRcode(ctx *gin.Context) {
	m := middleware.ExtractManeger(ctx)
	id := ctx.Param("session")

	session, ok := m.Get(id)
	if !ok {
		ctx.JSON(404, gin.H{
			"error": "Sessão não encontrada",
		})
		return
	}

	if session.Client.IsLoggedIn() {
		ctx.JSON(500, gin.H{"error": "você já está logado"})
		return
	}

	if !session.Client.IsConnected() {
		err := session.Start()
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	qr, err := session.GetQR()
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"base64":     qr.Base64,
		"expires_in": 20*60 - int(time.Since(session.LastQRTime).Seconds()),
	})

}
