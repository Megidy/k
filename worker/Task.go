package worker

import "github.com/Megidy/k/types"

func NewWorkerTask(TopicName string, players []types.Player) *WorkerTask {
	return &WorkerTask{
		TopicName: TopicName,
		Players:   players,
	}
}
