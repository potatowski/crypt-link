package main

import (
	"crypt-link/config"
	"crypt-link/database"
	"crypt-link/router"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.Load()
	fmt.Println("Configuration loaded successfully")

	err := database.Connect()
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	fmt.Println("Database connected successfully")

	r := router.Setup()
	fmt.Println("Router setup completed successfully")

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
