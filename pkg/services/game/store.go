package game

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/Megidy/k/types"
	"github.com/redis/go-redis/v9"
)

type store struct {
	sqlDB   *sql.DB
	redisDB *redis.Client
}

// TO DO
// remake question storage to mongoDB
// this should be perfect ngl

func NewGameStore(sqlDB *sql.DB, redisDB *redis.Client) *store {
	return &store{sqlDB: sqlDB, redisDB: redisDB}
}

func (s *store) CreateTopic(questions []types.Question, userID string) error {
	topic := questions[0].Topic
	_, err := s.sqlDB.Exec("INSERT INTO topics VALUES(?,?,?)", topic.TopicID, userID, topic.Name)
	if err != nil {
		log.Println("error occured in topics : ", err)
		return err

	}
	for _, question := range questions {

		_, err = s.sqlDB.Exec("INSERT INTO questions VALUES(?,?,?,?,?,?)", question.Id, question.Topic.TopicID, question.Type, question.ImageLink, question.Question, question.CorrectAnswer)
		if err != nil {
			log.Println("error occured in questions : ", err)
			return err
		}
		for _, answer := range question.Answers {
			_, err = s.sqlDB.Exec("INSERT INTO answers VALUES(?,?)", question.Id, answer)
			if err != nil {
				log.Println("error occured in answers : ", err)
				return err
			}
		}
	}
	topics, hasCache, err := s.GetCachedUsersTopics(userID)
	if err != nil {
		return err
	}
	if hasCache {
		topics = append(topics, *topic)
	}
	err = s.CacheTopics(userID, topics)
	if err != nil {
		return err
	}
	return nil
}

func (s *store) GetUsersTopics(userID string) ([]types.Topic, error) {
	var topics []types.Topic
	topics, hasCache, err := s.GetCachedUsersTopics(userID)
	if err != nil {
		return nil, err
	}
	if hasCache {
		log.Println("found cached topics : ", topics)
		return topics, nil
	}

	rows, err := s.sqlDB.Query("SELECT * FROM topics WHERE user_id=?", userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t types.Topic
		err = rows.Scan(&t.TopicID, &t.UserID, &t.Name)
		if err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	err = s.CacheTopics(userID, topics)
	if err != nil {
		return nil, err
	}
	return topics, nil
}

func (s *store) GetQuestionsByTopicName(TopicName string, userID string) ([]types.Question, error) {
	var questions []types.Question
	questions, hasCache, err := s.GetCachedUsersQuestions(userID, TopicName)
	if err != nil {
		return nil, err
	}
	if hasCache {
		return questions, nil
	}
	row, err := s.sqlDB.Query("SELECT * FROM topics WHERE name=? AND user_id=?", TopicName, userID)
	if err != nil {
		log.Println("error with topics")
		return nil, err
	}
	var t types.Topic
	for row.Next() {

		err = row.Scan(&t.TopicID, &t.UserID, &t.Name)
		if err != nil {
			return nil, err
		}
	}
	log.Println("TOPIC : ", t)
	rows, err := s.sqlDB.Query("SELECT id,type,image_link,question,correct_answer FROM questions WHERE topic_id=?", t.TopicID)
	if err != nil {
		log.Println("error with questions")
		return nil, err
	}
	for rows.Next() {
		var q types.Question
		q.Topic = &t
		// q.Answers = make([]string, 4)
		err = rows.Scan(&q.Id, &q.Type, &q.ImageLink, &q.Question, &q.CorrectAnswer)
		if err != nil {
			return nil, err
		}
		rows, err := s.sqlDB.Query("SELECT answer FROM answers WHERE question_id=?", q.Id)
		if err != nil {
			log.Println("error with answers ")
			return nil, err
		}

		for rows.Next() {
			var a string
			err = rows.Scan(&a)
			if err != nil {
				return nil, err
			}
			q.Answers = append(q.Answers, a)
		}
		questions = append(questions, q)
	}
	log.Println("questions to cache: ", questions)
	err = s.CacheQuestions(userID, questions)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *store) TopicNameAlreadyExists(userID string, topicName string) (bool, error) {
	rows, err := s.sqlDB.Query("SELECT 1 FROM topics WHERE user_id=? AND name=?", userID, topicName)
	if err != nil {
		return false, err
	}
	for !rows.Next() {
		return false, nil
	}
	return true, nil
}

func (s *store) GetCachedUsersTopics(userID string) ([]types.Topic, bool, error) {
	var topics []types.Topic
	ctx := context.Background()
	topicCacheKey := "user:" + userID + ":topics_name:"
	data, err := s.redisDB.Get(ctx, topicCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("no cache was found , querying from sql db")
		} else {
			return nil, false, err
		}
	} else {
		err := json.Unmarshal([]byte(data), &topics)
		if err != nil {
			return nil, false, err
		}
		log.Println("cached topics found : ", topics)
		return topics, true, nil
	}
	return nil, false, nil
}

func (s *store) GetCachedUsersQuestions(userID string, topicName string) ([]types.Question, bool, error) {
	ctx := context.Background()
	var questions []types.Question
	questionCacheKey := "user:" + userID + ":questions_of_topic:" + topicName
	data, err := s.redisDB.Get(ctx, questionCacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("no cache was found , querying from sql db")
		} else {
			return nil, false, err
		}
	} else {
		err := json.Unmarshal([]byte(data), &questions)
		if err != nil {
			return nil, false, err
		}
		log.Println("cached questions found : ", questions)
		return questions, true, nil
	}
	return nil, false, nil

}

func (s *store) CacheQuestions(userID string, questions []types.Question) error {
	ctx := context.Background()
	questionCacheKey := "user:" + userID + ":questions_of_topic:" + questions[0].Topic.Name
	log.Println("caching questions : ", questions)
	data, err := json.Marshal(questions)
	if err != nil {
		return err
	}
	err = s.redisDB.Set(ctx, questionCacheKey, data, time.Minute*5).Err()
	if err != nil {
		return err
	}
	return nil
}
func (s *store) CacheTopics(userID string, topics []types.Topic) error {
	ctx := context.Background()
	topicCacheKey := "user:" + userID + ":topics_name:"
	log.Println("caching topics : ", topics)
	data, err := json.Marshal(topics)
	if err != nil {
		return err
	}
	err = s.redisDB.Set(ctx, topicCacheKey, data, time.Minute*5).Err()
	if err != nil {
		return err
	}
	return nil
}
