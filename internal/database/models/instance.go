package models

import (
	"database/sql"
	"time"
)
type WebhookEvent int

type Instance struct {
	//Instacia
    ID        string   `gorm:"primaryKey;autoIncrement:false"`
	Name 	  string
	Token     string
	Number    sql.NullString `gorm:"type:text"`
	Status    int
	//Configurações
	RejectCall bool
	ReadMessages bool

	//webhook
	Webhook   string
	Webhook_events int
	//Outros
	CreatedAt time.Time
}


//estou usando bitmask semelhante ao discord
const (
	MessageReceived WebhookEvent = 1 << iota   	  // 1
	MessageSender							      // 2
	EventNewContact                               // 4
	EventQR                                       // 8
	EventLoggedIn                                 // 16
	EventLoggedOut                                // 32
	PairSuccess									  // 64 
)