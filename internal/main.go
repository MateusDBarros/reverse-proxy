package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
}
