package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	telegramAPIBase = "https://api.telegram.org"
)

var (
	validKey string
	port     string
)

func init() {
	validKey = os.Getenv("PROXY_KEY")
	if validKey == "" {
		log.Fatal("PROXY_KEY environment variable is required")
	}

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
}

func main() {
	http.HandleFunc("/", proxyHandler)

	log.Printf("Starting proxy server on port %s", port)
	log.Printf("Key validation enabled")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// Silently close connection - behave like port is closed
func dropConnection(w http.ResponseWriter) {
	if hj, ok := w.(http.Hijacker); ok {
		conn, _, _ := hj.Hijack()
		if conn != nil {
			conn.Close()
		}
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if !strings.HasPrefix(path, "/bot") {
		dropConnection(w)
		return
	}

	parts := strings.SplitN(path[4:], "/", 2)
	if len(parts) < 1 {
		dropConnection(w)
		return
	}

	keyToken := parts[0]
	method := ""
	if len(parts) > 1 {
		method = "/" + parts[1]
	}

	underscoreIndex := strings.Index(keyToken, "_")
	if underscoreIndex == -1 {
		dropConnection(w)
		return
	}

	providedKey := keyToken[:underscoreIndex]
	actualToken := keyToken[underscoreIndex+1:]

	if providedKey != validKey {
		log.Printf("Invalid key attempt: %s from %s", providedKey, r.RemoteAddr)
		dropConnection(w)
		return
	}

	targetURL := fmt.Sprintf("%s/bot%s%s", telegramAPIBase, actualToken, method)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Error executing proxy request: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response body: %v", err)
	}

	log.Printf("%s %s -> %d", r.Method, path, resp.StatusCode)
}
