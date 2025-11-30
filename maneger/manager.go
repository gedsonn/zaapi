package manager

var Manager = &SessionManager{
	Sessions: make(map[string]*Session),
}

func (m *SessionManager) Add(s *Session) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	// Adiciona na memória
	m.Sessions[s.ID] = s

	return nil
}

func (m *SessionManager) Get(id string) (*Session, bool) {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	s, ok := m.Sessions[id]
	return s, ok
}

func (m *SessionManager) Remove(id string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	delete(m.Sessions, id)
}

func NewManager() (*SessionManager, error) {
	m := NewEmptyManager()
	return m, nil
}

func NewEmptyManager() *SessionManager {
	return &SessionManager{
		Sessions: make(map[string]*Session),
	}
}

