package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type ServerPool struct {
	Servers []*Backend
	mutex   sync.RWMutex
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

func GetAttempts(r *http.Request) int {
	if attemps, ok := r.Context().Value("attempts").(int); ok {
		return attemps
	}
	return 0
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value("retry").(int); ok {
		return retry
	}
	return 0
}

type contextKey string

const Attempts contextKey = "attempts"

func (s *ServerPool) HttpHandler(w http.ResponseWriter, r *http.Request) {
	attempts, ok := r.Context().Value(Attempts).(int)
	if !ok {
		attempts = 1
	}

	if attempts > 3 {
		log.Printf("%s Max retries reached", r.URL.Path)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	peer := s.GetNextBackend()
	if peer == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	peer.IncrementConnections()
	// We don't use defer here because we are in a recursive-style call
	// Instead, we'll decrement when the proxy finishes or fails

	peer.proxy.ServeHTTP(w, r)
	peer.DecrementConnections()
}

func (s *ServerPool) HealthCheck() {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for _, b := range s.Servers {
		res, err := client.Get(b.url.String())
		if res != nil {
			res.Body.Close()
		}
		alive := err == nil && res.StatusCode == http.StatusOK

		b.SetAlive(alive)

	}
}
