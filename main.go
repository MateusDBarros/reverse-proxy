package main

import (
	"net/http/httputil"
	"sync"
)

type Backend struct {
	url               string
	alive             bool
	activeConnections int
	mu                sync.RWMutex
	proxy             *httputil.ReverseProxy
}

type BackendFunc interface {
	IsAlive() bool
	SetAlive(alive bool)
	IncrementConnections()
	DecrementConnections()
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.alive
}

func (b *Backend) SetAlive(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.alive = healthy
}

func (b *Backend) IncrementConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.activeConnections++
}

func (b *Backend) DecrementConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.activeConnections--
}

type ServerPool struct {
	Servers []*Backend
	mutex   sync.RWMutex
	MaxConn int
	counter int
}

func (s *ServerPool) IncrementCounter() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.counter++
}

func (s *ServerPool) AddBackend(b *Backend) *Backend {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Servers = append(s.Servers, b)
	b.alive = true
	return b
}

func (s *ServerPool) GetNextBackend() *Backend {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.Servers) == 0 {
		return nil
	}

	for i := 0; i < len(s.Servers); i++ {
		start := s.counter
		target := (start + 1 + i) % len(s.Servers)
		if s.Servers[target].alive {
			s.counter = (target + 1) % len(s.Servers)
			return s.Servers[target]
		}
	}
	return nil
}
