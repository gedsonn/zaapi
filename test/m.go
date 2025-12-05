package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/apex/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type Chat struct {
	Number string `json:"phone"`
	Name   string `json:"name"`
	FromMe bool   `json:"fromMe"`
}

type Test int

const (
	None Test = 0
	Text Test = 1 << iota
	Image
	Video
)

// Media represents a media file (image, video, etc.).
type Media struct {
	Identify      string `json:"url"`
	Mimetype string `json:"mimetype"`
}

type CleanMessage struct {
	Chat  Chat   `json:"chat"`
	Text  string `json:"text,omitempty"`
	Type  Test   `json:"type"`
	Media *Media `json:"media,omitempty"`
}

func ConvertReceived(evt *events.Message, client *whatsmeow.Client) CleanMessage {
	m := CleanMessage{
		Chat: Chat{
			Number: evt.Info.Chat.User,
			Name:   evt.Info.PushName,
			FromMe: evt.Info.IsFromMe,
		},
		Type: None,
	}

	if text := evt.Message.GetConversation(); text != "" {
		m.Text = text
		m.Type |= Text
	} else if extendedText := evt.Message.ExtendedTextMessage.GetText(); extendedText != "" {
		m.Text = extendedText
		m.Type |= Text
	}

	if img := evt.Message.ImageMessage; img != nil {
		cwd, err := os.Getwd()
		if err != nil {
			log.Warn("Falha ao obter diretório atual, usando fallback em /etc/zaapi/config.yml")
		}
		mid := time.Now().UnixNano()
		fileName := fmt.Sprintf("%d.jpg", mid)
		filePath := filepath.Join(fmt.Sprintf("%s/assets/images", cwd), fileName)
		m.Type |= Image
		if caption := img.GetCaption(); caption != "" {
			m.Text = caption
			m.Type |= Text
		}

		data, err := client.Download(context.Background(), img)
		if err != nil {
			fmt.Println("Erro ao baixar imagem:", err)
			return m
		}

		err = os.WriteFile(filePath, data, os.ModePerm)
		if err != nil {
			fmt.Println("Erro ao salvar imagem:", err)
			return m
		}

		fmt.Println("Imagem salva em:", filePath)

		// apaga após 10 minutos
		time.AfterFunc(2*time.Minute, func() {
			os.Remove(filePath)
			fmt.Println("Imagem removida:", filePath)
		})
		if path := img.GetDirectPath(); path != "" {
			m.Media = &Media{
				Identify:      strconv.Itoa(int(mid)),
				Mimetype: img.GetMimetype(),
			}
		}
	}

	if vid := evt.Message.VideoMessage; vid != nil {
		m.Type |= Video
		if caption := vid.GetCaption(); caption != "" {
			m.Text = caption
			m.Type |= Text
		}
		if path := vid.GetDirectPath(); path != "" {
			m.Media = &Media{
				Identify:     path,
				Mimetype: vid.GetMimetype(),
			}
		}
	}

	return m
}
