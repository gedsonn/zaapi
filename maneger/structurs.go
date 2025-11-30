package manager

import (
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
)

type QRCodeEvent struct {
	Code   string
	Base64 string
	Event  string // "code", "timeout", "success"
}

type Session struct {
	ID       string
	Path     string
	Client   *whatsmeow.Client
	StopChan chan struct{} // sinal para parar de escutar
	// EventHandlers permite funções customizadas que serão chamadas quando
	// um evento do WhatsMeow for recebido. A chave é um identificador
	// retornado por AddEventHandler e usado para remoção.
	EventHandlers map[int]func(interface{})
	// HandlerIDCounter é um contador para gerar IDs únicos para handlers.
	HandlerIDCounter int64
	Stopped          atomic.Bool // estado da sessão
	QRChan           chan QRCodeEvent
	listenerStarted  atomic.Bool
	LastQR           QRCodeEvent
	LastQRTime       time.Time
	Mu	             sync.RWMutex
	Ready            bool
	LoggedIn         bool
}

type SessionManager struct {
	Sessions map[string]*Session
	Mutex    sync.RWMutex
}

type Credentials struct {
	Token string `yaml:"token"`
	Client string `yaml:"client"`
}

type SessionYml struct {
	Id       string `yaml:"id"`
	Status 	 string `yaml:"status"`
	Credentials Credentials `yaml:"credentials"`
}
