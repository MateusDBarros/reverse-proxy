package main

import (
	"net/http"
	"sync"
	"time"
)

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
	start := s.counter
	for i := 0; i < len(s.Servers); i++ {

		target := (start + 1 + i) % len(s.Servers)
		if s.Servers[target].alive {
			s.counter = (target + 1) % len(s.Servers)
			return s.Servers[target]
		}
	}
	return nil
}

func (s *ServerPool) HttpHandler(w http.ResponseWriter, r *http.Request) {

	peer := s.GetNextBackend()
	if peer == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	peer.IncrementConnections()
	defer peer.DecrementConnections()
	peer.proxy.ServeHTTP(w, r)
}

func (s *ServerPool) HealthCheck() {
	// 1. Create a client with a timeout
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for _, b := range s.Servers {
		res, err := client.Get(b.url)

		alive := err == nil && res.StatusCode == http.StatusOK

		b.SetAlive(alive)

		if res != nil {
			res.Body.Close()
		}
	}
}
