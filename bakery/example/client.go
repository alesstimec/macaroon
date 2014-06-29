package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/rogpeppe/macaroon/httpbakery"
)

func client(serverEndpoint string) {
	req, err := http.NewRequest("GET", serverEndpoint, nil)
	if err != nil {
		log.Fatalf("new request error: %v", err)
	}
	resp, err := httpbakery.Do(http.DefaultClient, req)
	if err != nil {
		log.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("GET succeeded")
	io.Copy(os.Stdout, resp.Body)
}
