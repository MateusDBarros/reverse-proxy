package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
	// any other imports you need
)

func main() {

	pool := ServerPool{}

	addresses := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082", "http://localhost:8083"}

	for _, addr := range addresses {

		parsedUrl, err := url.Parse(addr)
		if err != nil {
			log.Fatal(err)
		}

		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(parsedUrl)
				pr.SetXForwarded()
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {

				log.Printf("Proxy error: %v", err)
			},
		}

		backend := &Backend{url: parsedUrl, alive: true, proxy: proxy}
		pool.AddBackend(backend)

	}
	go func() {
		for {
			pool.HealthCheck()
			time.Sleep(10 * time.Second)
		}
	}()

	// 5. Start the Load Balancer
	log.Println("Load Balancer starting on port :9420...")

	// Remember to wrap your pool.HttpHandler in http.HandlerFunc() !
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pool.HttpHandler(w, r)
	})
	if err := http.ListenAndServe(":9420", nil); err != nil {
		log.Fatal(err)
	}
}
