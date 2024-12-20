package main

import (
	"log"

	"github.com/Megidy/k/api"
	"github.com/Megidy/k/db"
)

func main() {
	db, err := db.NewDB()
	if err != nil {
		log.Fatalln("error when establishing connection to db : ", err)
	}
	log.Println("started DB successfully")
	server := api.NewServer(":8080", db)
	err = server.Run()
	if err != nil {
		log.Fatalln("error while hosting server :", err)
	}
}
