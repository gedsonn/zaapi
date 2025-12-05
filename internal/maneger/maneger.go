package maneger

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"

	"github.com/goccy/go-yaml"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)


type Manager struct {
	Instacias map[string]*Instancia
	Mu        sync.Mutex
}

var Maneger = &Manager{
	Instacias: make(map[string]*Instancia),
	Mu:        sync.Mutex{},
}

func Load() (*Manager, error) {
	m := EmptyManager()
	return m, nil
}

func EmptyManager() *Manager {
	return &Manager{
		Instacias: make(map[string]*Instancia),
		Mu:        sync.Mutex{},
	}
}

// adicionar Instancia na memoria
func (m *Manager) Add(s *Instancia) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Instacias[s.Id] = s

	return nil
}

// pegar Instancia na memoria
func (m *Manager) Get(id string) (*Instancia, bool) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	s, ok := m.Instacias[id]
	return s, ok
}

func (m *Manager) Remove(id string) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	delete(m.Instacias, id)
}

func (m *Manager) Sync() error {
	dirs, err := os.ReadDir("sessions")
	if err != nil {
		return err
	}

	

	for _, d := range dirs {
		
		if !d.IsDir() {
			continue
		}

		fmt.Printf("%v", d)

		id := d.Name()

		path := fmt.Sprintf("sessions/%s/session.yml", id)

		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return err
		}

		var s InstaciaYml
		if err := yaml.Unmarshal(data, &s); err != nil {
			return err
		}

		
	}

	return nil
}

func (m *Manager) RestoreInstance(id string) (*Instancia, error) {
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

