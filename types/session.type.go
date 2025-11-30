package types

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

type WASession struct {
    Client *whatsmeow.Client
    DeviceStore *sqlstore.Container
}

var SessionManager = map[string]*WASession{}
