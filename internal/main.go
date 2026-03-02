package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {

	pool := ServerPool{}

	addresses := [3]string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"}

	for _, address := range addresses {
		targetURL, _ := url.Parse(address)

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		var b *Backend

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("Proxy Error: %v", e)

			b.SetAlive(false)

			attempts, _ := r.Context().Value(Attempts).(int)
			if attempts == 0 {
				attempts = 1
			}

			ctx := context.WithValue(r.Context(), Attempts, attempts+1)

			pool.HttpHandler(w, r.WithContext(ctx))
		}

		b = &Backend{
			url:   targetURL,
			proxy: proxy,
			alive: true,
		}
		pool.AddBackend(b)
	}

	go func() {
		for {
			log.Println("Performing health check")
			pool.HealthCheck()
			time.Sleep(10 * time.Second)
		}
	}()

	err := http.ListenAndServe(":9420", http.HandlerFunc(pool.HttpHandler))
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
