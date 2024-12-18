package main

import (
	"log"

	"github.com/Megidy/k/api"
)

func main() {
	server := api.NewServer(":8080")
	err := server.Run()
	if err != nil {
		log.Fatalln("error while hosting server :", err)
	}
}
