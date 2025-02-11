package worker

import "github.com/Megidy/k/types"

type WorkerManager interface {
	Run()
	Worker()
	StartTrackingGame(task *WorkerTask)
}

type WorkerTask struct {
	TopicName string
	Players   []types.Player
}
