package storage

import (
	"bytes"
	"context"
	"io"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string][]byte)}
}

func (m *MemoryStorage) Upload(_ context.Context, key, _ string, r io.Reader) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.data[key] = body
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	body, ok := m.data[key]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}

func (m *MemoryStorage) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.data, key)
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) PublicURL(key string) string {
	return "memory://" + key
}
