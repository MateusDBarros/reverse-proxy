package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	url               *url.URL
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
