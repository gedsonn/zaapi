package maneger

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apex/log"
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

// QRCodeEvent representa os dados de um evento de QR code para login.
type QRCodeEvent struct {
	Code   string
	Base64 string
}

// Instancia gerencia uma única sessão/conexão com o WhatsApp.
type Instancia struct {
	Id     string
	Client *whatsmeow.Client
	Mu     sync.RWMutex // Protege o acesso concorrente à instância

	// Flags atômicas para um estado seguro entre goroutines.
	Stopped atomic.Bool // Se true, a instância está parada.
	Listen  atomic.Bool // Se true, o handler de eventos está registrado.

	// Armazena o último QR code gerado e o tempo de geração.
	LastQR     *QRCodeEvent
	LastQRTime time.Time
}

type InstaciaYml struct {
	Id      string
	Listen  bool
	Stopped bool
}

// Start inicia a conexão da instância com o WhatsApp.
func (i *Instancia) Start() error {
	i.Mu.Lock()
	defer i.Mu.Unlock()

	if !i.Stopped.Load() {
		return fmt.Errorf("instancia já está online")
	}

	i.Stopped.Store(false)
	i.ensureListener()

	// Se o cliente não estiver logado, inicia o fluxo de conexão via QR code.
	// Caso contrário, apenas conecta.
	if i.Client.Store.ID == nil {
		go i.qrConnectionFlow()
	} else {
		if err := i.Client.Connect(); err != nil {
			return err
		}
	}

	return nil
}

// qrConnectionFlow gerencia o ciclo de vida da conexão por QR code.
// Esta função é executada em uma goroutine para não bloquear.
func (i *Instancia) qrConnectionFlow() {
	// Se já estiver conectado ou logado, não faz nada.
	if i.Client.IsConnected() && i.Client.IsLoggedIn() {
		return
	}

	// 1. Solicita o canal de eventos de QR code ANTES de conectar.
	// A biblioteca whatsmeow exige que GetQRChannel seja chamado antes de Connect.
	qrChan, err := i.Client.GetQRChannel(context.Background())
	if err != nil {
		log.Errorf("Erro ao obter QR channel: %v", err)
		return
	}

	// 2. Conecta ao WhatsApp.
	// Esta chamada irá bloquear até que a conexão seja estabelecida ou falhe.
	if err := i.Client.Connect(); err != nil {
		log.Errorf("Erro ao conectar: %v", err)
		return
	}

	// 3. Processa os eventos do canal de QR code em uma nova goroutine
	// para não bloquear o fluxo principal da conexão.
	for evt := range qrChan {
		switch evt.Event {
		case "code":
			// Novo QR code gerado.
			png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			if err != nil {
				log.Errorf("Erro ao encodar QR code para PNG: %v", err)
				continue
			}
			encoded := base64.StdEncoding.EncodeToString(png)

			i.Mu.Lock()
			i.LastQR = &QRCodeEvent{Code: evt.Code, Base64: encoded}
			i.LastQRTime = time.Now()
			i.Mu.Unlock()

		case "timeout", "error", "success":
			// O fluxo de QR terminou (expirou, deu erro ou teve sucesso).
			// A goroutine será encerrada pois o canal será fechado pela biblioteca.
			log.Infof("Evento de login: %s", evt.Event)
			i.Mu.Lock()
			i.LastQR = nil // Limpa o QR code antigo.
			i.Mu.Unlock()
			return // Encerra a goroutine.
		}
	}
}

// Stop para a instância e desconecta o cliente do WhatsApp.
func (i *Instancia) Stop() error {
	i.Mu.Lock()
	defer i.Mu.Unlock()

	if i.Stopped.Load() {
		return nil
	}

	i.Stopped.Store(true)

	// Desconecta client
	if i.Client != nil && i.Client.IsConnected() {
		i.Client.Disconnect()
	}

	return nil
}

// ensureListener garante que o handler de eventos seja registrado apenas uma vez.
func (i *Instancia) ensureListener() {
	if i.Listen.Load() {
		return
	}
	go i.Client.AddEventHandler(i.handleEvent)
	i.Listen.Store(true)
}

// handleEvent processa eventos recebidos do cliente whatsmeow.
func (i *Instancia) handleEvent(evt interface{}) {
	if i.Stopped.Load() || !i.Listen.Load() {
		return
	}

	switch e := evt.(type) {
	case *events.Message:

	case *events.PairSuccess:
		fmt.Printf("%v", e)

	default:

	}
}

// GetQR retorna o QR code mais recente para login.
// Se nenhum QR code estiver disponível ou se estiver expirado, ele tenta iniciar um novo fluxo.
func (i *Instancia) GetQR() (*QRCodeEvent, error) {
	if i.Stopped.Load() {
		return nil, fmt.Errorf("instância(%s) está parada", i.Id)
	}

	if i.Client.IsLoggedIn() {
		return nil, fmt.Errorf("sessão já está conectada")
	}

	i.Mu.RLock()
	lastQR := i.LastQR
	i.Mu.RUnlock()

	// Se não houver QR code, ou se o fluxo de QR não estiver ativo, inicia um novo.
	// A biblioteca fecha o canal de QR em eventos como 'timeout', então precisamos de um novo.
	if lastQR == nil {
		go i.qrConnectionFlow()
		return nil, fmt.Errorf("solicitando novo QR code, tente novamente em alguns segundos")
	}

	return lastQR, nil
}

func CreateInstance(id string) (*Instancia, error) {
	if len(id) < 1 {
		return nil, fmt.Errorf("id invalido")
	}

	ctx := context.Background()
	path := fmt.Sprintf("sessions/%s", id)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	dbpath := fmt.Sprintf("file:%s/store.db?_foreign_keys=on", path)
	container, err := sqlstore.New(ctx, "sqlite3", dbpath, waLog.Noop)
	if err != nil {
		return nil, err
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	// Identidade do aparelho
	store.DeviceProps = &waCompanionReg.DeviceProps{
		Os:           proto.String("Windows"),                  //windowns
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

	Instance := &Instancia{
		Id:      id,
		Client:  client,
		Mu:      sync.RWMutex{},
		Stopped: atomic.Bool{},
		Listen:  atomic.Bool{},
	}

	Instance.Stopped.Store(true)
	Instance.Listen.Store(false)

	return Instance, nil
}
