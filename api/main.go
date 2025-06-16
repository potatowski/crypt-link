package main

import (
	"crypt-link/config"
	"crypt-link/router"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	r := router.Setup(cfg)

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
