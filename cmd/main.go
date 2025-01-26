package main

import (
	"log"
	"runtime"
	"time"

	"github.com/Megidy/k/api"
	"github.com/Megidy/k/db"
)

func main() {
	go func() {
		for {
			log.Println("number of goruotines : ", runtime.NumGoroutine())
			time.Sleep(time.Second * 2)
		}
	}()
	sqlDB, err := db.NewSQlDB()
	if err != nil {
		log.Fatalln("error when establishing connection to sql db : ", err)
	}
	log.Println("started SQL DB successfully")
	redisDB, err := db.NewRedisDB()
	if err != nil {
		log.Fatalln("error when establishing connection to redis db : ", err)
	}
	log.Println("started redis DB successfully")
	server := api.NewServer(":8080", sqlDB, redisDB)
	err = server.Run()
	if err != nil {
		log.Fatalln("error while hosting server :", err)
	}

}
