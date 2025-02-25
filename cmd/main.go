package main

import (
	"log"

	"github.com/Megidy/k/cmd/api"
	"github.com/Megidy/k/config"
	"github.com/Megidy/k/db"
)

func main() {
	cfg := config.NewConfig()

	sqlDB, err := db.NewSQlDB(cfg)
	if err != nil {
		log.Fatalln("error when establishing connection to sql db : ", err)
	}
	log.Println("started SQL DB successfully")
	redisDB, err := db.NewRedisDB()
	if err != nil {
		log.Fatalln("error when establishing connection to redis db : ", err)
	}
	log.Println("started redis DB successfully")
	server := api.NewServer(cfg, ":8080", sqlDB, redisDB)
	err = server.Run()
	if err != nil {
		log.Fatalln("error while hosting server :", err)
	}

}
