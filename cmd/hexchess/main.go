package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	hexchess "github.com/Mixel-MurP-427/hex_chess"
)

func main() {
	address := flag.String("addr", "127.0.0.1:8072", "local address to serve the game on")
	flag.Parse()

	server := &http.Server{
		Addr:              *address,
		Handler:           hexchess.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Hex Chess is available at http://%s", *address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve Hex Chess: %v", err)
	}
}
