package state

import (
	"sync"

	"linux-hardener/internal/policy"
)

type AppState struct {
	mu     sync.RWMutex
	Policy policy.Policy
	Dirty  bool
}

var global = &AppState{}

func Global() *AppState { return global }

func (s *AppState) SetPolicy(p policy.Policy) {
	s.mu.Lock()
	s.Policy = p
	s.Dirty = false
	s.mu.Unlock()
}

func (s *AppState) GetPolicy() policy.Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Policy
}

func (s *AppState) MarkDirty() {
	s.mu.Lock()
	s.Dirty = true
	s.mu.Unlock()
}
