package main

import "net/http"

func lbHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the next backend from the pool
	// 2. If it exists, call: backend.proxy.ServeHTTP(w, r)
}
