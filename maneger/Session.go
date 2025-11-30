package manager

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"encoding/base64"
	"encoding/json"

	"github.com/goccy/go-yaml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

//criar nova instancia
func (m *SessionManager) CreateSession(id string) (*Session, error) {
	ctx := context.Background()
	path := fmt.Sprintf("sessions/%s", id)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	db := fmt.Sprintf("file:%s/store.db?_foreign_keys=on", path)
	container, err := sqlstore.New(ctx, "sqlite3", db, waLog.Noop)
	if err != nil {
		return nil, err
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	// Identidade do aparelho
	store.DeviceProps = &waCompanionReg.DeviceProps{
		Os:           proto.String("Windows"), //windown
		PlatformType: waCompanionReg.DeviceProps_CHROME.Enum(), //chrome
		Version: &waCompanionReg.DeviceProps_AppVersion{
			Primary:   proto.Uint32(118),
			Secondary: proto.Uint32(0),
			Tertiary:  proto.Uint32(2),
		},
		HistorySyncConfig: &waCompanionReg.DeviceProps_HistorySyncConfig{
			SupportGroupHistory:   proto.Bool(false),
			StorageQuotaMb:        proto.Uint32(0),
			RecentSyncDaysLimit:   proto.Uint32(0), //não fazer sync
			SupportCallLogHistory: proto.Bool(false),
		},
	}

	client := whatsmeow.NewClient(device, waLog.Noop)

	session := &Session{
		ID:       id,
		Client:   client,
		LoggedIn: client.Store.ID != nil,
		Ready:    false,
		Mu:       sync.RWMutex{},
		Stopped:  atomic.Bool{},
		StopChan: make(chan struct{}),
		QRChan:   make(chan QRCodeEvent, 1),
	}

	session.Stopped.Store(true)
	session.listenerStarted.Store(false)

	err = m.Add(session)
	if err != nil {
		return nil, err
	}

	d := SessionYml{
		Id:     id,
		Status: "running",
		Credentials: Credentials{
			Token:  "123",
			Client: "123",
		},
	}
	data, err := yaml.Marshal(&d)
	if err != nil {
		return nil, err
	}

	ymlFile := fmt.Sprintf("sessions/%s/session.yml", id)

	err = os.WriteFile(ymlFile, data, 0644)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Session) Start() error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if !s.Stopped.Load() {
		return fmt.Errorf("session already started")
	}

	s.Stopped.Store(false)
	s.StopChan = make(chan struct{}) // novo canal
	s.ensureListener()

	// Conectar client (simulação)
	if err := s.Client.Connect(); err != nil {
		s.Stopped.Store(true)
		close(s.StopChan)
		return fmt.Errorf("connect error: %w", err)
	}

	return nil
}

// --- Stop da sessão ---
func (s *Session) Stop() {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.Stopped.Load() {
		return
	}

	s.Stopped.Store(true)

	// Fecha canal para parar listener
	if s.StopChan != nil {
		select {
		case <-s.StopChan:
			// já fechado
		default:
			close(s.StopChan)
		}
		s.StopChan = nil
	}

	// Desconecta client
	if s.Client != nil && s.Client.IsConnected() {
		s.Client.Disconnect()
	}
	s.Ready = false
}

// --- Listener ---
func (s *Session) ensureListener() {
	if s.listenerStarted.Load() {
		return
	}
	go s.registerEventHandler()
	s.listenerStarted.Store(true)
}

func (s *Session) registerEventHandler() {
	s.Client.AddEventHandler(s.handleEvent)
}

func (s *Session) handleEvent(evt interface{}) {
	//ignora caso a instacia esteja parado
	if s.Stopped.Load() {
        return
    }
	
	switch e := evt.(type) {
	//debug de messagens
	case *events.Message:
		data, err := json.MarshalIndent(e.Message, "", "  ")
        if err != nil {
            fmt.Println("Erro ao converter mensagem para JSON:", err)
            return
        }
        fmt.Printf("[msg] session=%s message=%s\n", s.ID, string(data))
	case *events.PairSuccess:
		s.Mu.Lock()
		s.LoggedIn = true
		s.Mu.Unlock()
		fmt.Printf("numero: %s\n", e.ID.User)
		
	
	default:
		fmt.Printf("[evt] session=%s type=%T\n", s.ID, evt)
	}
}

func (s *Session) GetQR() (QRCodeEvent, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// Se existe QR salvo e não expirou
	if s.LastQR.Code != "" && time.Since(s.LastQRTime) < 1*time.Minute {
		return s.LastQR, nil
	}

	// QR expirou — gerar novo
	return s.generateNewQR()
}

func (s *Session) generateNewQR() (QRCodeEvent, error) {
	ctx := context.Background()

	// SE estiver conectado, desconectar limpo
	if s.Client.IsConnected() {
		s.Client.Disconnect()
	}

	// Criar canal de QR
	qrChan, _ := s.Client.GetQRChannel(ctx)

	// Conectar
	if err := s.Client.Connect(); err != nil {
		return QRCodeEvent{}, err
	}

	// Ler somente o PRIMEIRO QR válido
	for {
		qr, ok := <-qrChan
		if !ok {
			return QRCodeEvent{}, fmt.Errorf("qr channel closed")
		}

		if qr.Event == "code" {
			// Converter QR → PNG
			png, err := qrcode.Encode(qr.Code, qrcode.Medium, 256)
			if err != nil {
				return QRCodeEvent{}, err
			}

			// PNG → Base64
			encoded := base64.StdEncoding.EncodeToString(png)

			e := QRCodeEvent{
				Code:   qr.Code,
				Base64: encoded,
				Event:  "code",
			}

			// Salva QR na sessão
			s.LastQR = e
			s.LastQRTime = time.Now()

			return e, nil
		}

		// Se for "timeout", "error", etc, continuar
	}
}
