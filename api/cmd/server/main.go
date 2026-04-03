package main

import (
	"crypt-link/adapter/input/controller"
	"crypt-link/adapter/output/mongodb"
	"crypt-link/config"
	"crypt-link/core/service"
	"crypt-link/database"
	"crypt-link/router"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.Load()
	fmt.Println("Configuration loaded successfully")

	dbClient, err := database.Connect()
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	fmt.Println("Database connected successfully")

	// Dependency Injection (Hexagonal / Clean Architecture)
	repo := mongodb.NewMessageRepository(dbClient)
	svc := service.NewMessageService(repo)
	ctrl := controller.NewMessageController(svc)

	r := router.Setup(ctrl)
	fmt.Println("Router setup completed successfully")

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
