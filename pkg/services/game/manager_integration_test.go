package game

import (
	"testing"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
)

type MockGameStore struct{}

func (s *MockGameStore) CreateTopic(questions []types.Question, userID string) error {
	return nil
}
func (s *MockGameStore) GetUsersTopics(userID string) ([]types.Topic, error) {
	return nil, nil
}
func (s *MockGameStore) GetQuestionsByTopicName(TopicName string, userID string) ([]types.Question, error) {
	return nil, nil
}
func (s *MockGameStore) TopicNameAlreadyExists(userID string, topicName string) (bool, error) {
	return true, nil
}
func (s *MockGameStore) GetCachedUsersTopics(userID string) ([]types.Topic, bool, error) {
	return nil, false, nil
}
func (s *MockGameStore) GetCachedUsersQuestions(userID string, topicName string) ([]types.Question, bool, error) {
	return nil, false, nil
}
func (s *MockGameStore) CacheQuestions(userID string, questions []types.Question) error {
	return nil
}
func (s *MockGameStore) CacheTopics(userID string, topics []types.Topic) error {
	return nil
}

type MockClientSideHandler struct{}

func (h *MockClientSideHandler) GetDataFromForm(c *gin.Context, key string) string {
	return ""
}

func TestGame(t *testing.T) {
	t.Run("Should run if defalut flow of game is correct", func(t *testing.T) {
		// clientSideHandler := &MockClientSideHandler{}
		// gameStore := &MockGameStore{}
		// workerPool := worker.NewMockedWorkerPool()
		// _ := NewGameHandler(config.NewMockConfig(20, 30), clientSideHandler, gameStore, workerPool)

	})
}
