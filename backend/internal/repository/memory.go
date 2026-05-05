package repository

import (
	"sync"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/model"
)

type MemoryRepository struct {
	mu       sync.RWMutex
	users    map[string]model.User    // id → user
	byEmail  map[string]string        // email → id
	sessions map[string]model.Session // id → session
	byToken  map[string]string        // tokenHash → id
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:    make(map[string]model.User),
		byEmail:  make(map[string]string),
		sessions: make(map[string]model.Session),
		byToken:  make(map[string]string),
	}
}

func (m *MemoryRepository) CreateUser(u model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byEmail[u.Email]; exists {
		return ErrConflict
	}
	m.users[u.ID] = u
	m.byEmail[u.Email] = u.ID
	return nil
}

func (m *MemoryRepository) GetUserByID(id string) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return model.User{}, ErrNotFound
	}
	return u, nil
}

func (m *MemoryRepository) GetUserByEmail(email string) (model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byEmail[email]
	if !ok {
		return model.User{}, ErrNotFound
	}
	return m.users[id], nil
}

func (m *MemoryRepository) ListUsers(limit, offset int) ([]model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]model.User, 0, len(m.users))
	for _, u := range m.users {
		all = append(all, u)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MemoryRepository) CreateSession(s model.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID] = s
	m.byToken[s.TokenHash] = s.ID
	return nil
}

func (m *MemoryRepository) GetSessionByTokenHash(hash string) (model.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byToken[hash]
	if !ok {
		return model.Session{}, ErrNotFound
	}
	s := m.sessions[id]
	if s.ExpiresAt.Before(time.Now()) {
		return model.Session{}, ErrNotFound
	}
	return s, nil
}

func (m *MemoryRepository) DeleteSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.byToken, s.TokenHash)
	delete(m.sessions, id)
	return nil
}

func (m *MemoryRepository) DeleteSessionsByUserID(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.sessions {
		if s.UserID == userID {
			delete(m.byToken, s.TokenHash)
			delete(m.sessions, id)
		}
	}
	return nil
}
