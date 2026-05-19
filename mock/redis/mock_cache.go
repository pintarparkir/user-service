// Package redis provides an in-memory mock of the redis.Collections interface
// for use in unit tests that exercise cache-aware code paths.
package redis

import (
	"context"
	"sync"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

// MockCache is a simple thread-safe in-memory cache that implements redis.Collections.
type MockCache struct {
	mu    sync.Mutex
	Store map[string]string // exported so tests can inspect / pre-seed it
}

// NewMockCache returns an empty mock cache.
func NewMockCache() *MockCache {
	return &MockCache{Store: map[string]string{}}
}

func (m *MockCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch v := value.(type) {
	case string:
		m.Store[key] = v
	case []byte:
		m.Store[key] = string(v)
	}
	return nil
}

func (m *MockCache) Get(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.Store[key]
	if !ok {
		return "", goredis.Nil
	}
	return v, nil
}

func (m *MockCache) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Store[key]; ok {
		return false, nil
	}
	switch v := value.(type) {
	case string:
		m.Store[key] = v
	case []byte:
		m.Store[key] = string(v)
	}
	return true, nil
}

func (m *MockCache) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.Store, k)
	}
	return nil
}

func (m *MockCache) Eval(_ context.Context, _ string, _ []string, _ ...interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockCache) Ping(_ context.Context) error { return nil }

// Raw returns nil — the mock does not wrap a real redis client.
// Tests must not call Raw().
func (m *MockCache) Raw() *goredis.Client { return nil }
