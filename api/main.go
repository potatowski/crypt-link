package main

import (
	"crypt-link/config"
	"crypt-link/router"
	"log"
	"net/http"
)

func main() {
	config.Load()

	r := router.Setup()

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
