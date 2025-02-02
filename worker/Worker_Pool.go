package worker

import (
	"log"
	"strconv"

	"github.com/Megidy/k/types"
)

type WorkerPool struct {
	UserStore       types.UserStore
	TaskQueue       chan *WorkerTask
	NumberOfWorkers int
}

func NewWorkerPool(userStore types.UserStore, NumberOfWorkers int) *WorkerPool {
	return &WorkerPool{
		UserStore:       userStore,
		TaskQueue:       make(chan *WorkerTask, 100),
		NumberOfWorkers: NumberOfWorkers,
	}

}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.NumberOfWorkers; i++ {
		go wp.Worker()
	}
}

func (wp *WorkerPool) Worker() {
	for task := range wp.TaskQueue {
		log.Println("started Processing game with this leaderBoard : ", task.Players)
		for place, player := range task.Players {
			score := strconv.Itoa(player.Score)
			strPlace := strconv.Itoa(place + 1)
			err := types.UserStore.CacheUserGameScore(wp.UserStore, player.Username, score, strPlace, task.TopicName)
			if err != nil {
				log.Println("error : ", err)
			}
			log.Println("cached for : ", player.Username)
		}
		log.Println("ended up caching")

	}
}

func (wp *WorkerPool) StartTrackingGame(task *WorkerTask) {
	wp.TaskQueue <- task
}
