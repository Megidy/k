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

func NewGameStore(sqlDB *sql.DB, redisDB *redis.Client) *store {
	return &store{sqlDB: sqlDB, redisDB: redisDB}
}

func (s *store) CreateTopic(questions []types.Question, userID string) error {
	topic := questions[0].Topic
	_, err := s.sqlDB.Exec("insert into topics values(?,?,?)", topic.TopicID, userID, topic.Name)
	if err != nil {
		log.Println("error occured in topics : ", err)
		return err

	}
	for _, question := range questions {

		_, err = s.sqlDB.Exec("insert into questions values(?,?,?,?,?,?)", question.Id, question.Topic.TopicID, question.Type, question.ImageLink, question.Question, question.CorrectAnswer)
		if err != nil {
			log.Println("error occured in questions : ", err)
			return err
		}
		for _, answer := range question.Answers {
			_, err = s.sqlDB.Exec("insert into answers values(?,?)", question.Id, answer)
			if err != nil {
				log.Println("error occured in answers : ", err)
				return err
			}
		}
	}

	return nil
}

func (s *store) GetUsersTopics(userID string) ([]types.Topic, error) {
	var topics []types.Topic
	ctx := context.Background()
	cacheKey := "user:" + userID + ":topics_name:"
	data, err := s.redisDB.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("no cache was found , querying from sql db")
		} else {
			return nil, err
		}
	} else {
		err := json.Unmarshal([]byte(data), &topics)
		if err != nil {
			return nil, err
		}
		log.Println("cache found")
		return topics, nil
	}
	rows, err := s.sqlDB.Query("select * from topics where user_id=?", userID)
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
	if len(topics) > 0 {
		data, err := json.Marshal(topics)
		if err != nil {
			return nil, err
		}
		err = s.redisDB.Set(ctx, cacheKey, data, time.Minute*10).Err()
		if err != nil {
			return nil, err
		}
	}
	return topics, nil
}

func (s *store) GetQuestionsByTopicName(TopicName string, userID string) ([]types.Question, error) {
	var questions []types.Question
	ctx := context.Background()
	cacheKey := "user:" + userID + ":questions_of_topic:" + TopicName
	data, err := s.redisDB.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("no cache was found , querying from sql db")
		} else {
			return nil, err
		}
	} else {
		err := json.Unmarshal([]byte(data), &questions)
		if err != nil {
			return nil, err
		}
		log.Println("cache found")
		return questions, nil
	}
	row, err := s.sqlDB.Query("select * from topics where name=? and user_id=?", TopicName, userID)
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

	rows, err := s.sqlDB.Query("select id,type,image_link,question,correct_answer from questions where topic_id=?", t.TopicID)
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
		rows, err := s.sqlDB.Query("select answer from answers where question_id=?", q.Id)
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
	if len(questions) > 0 {
		data, err := json.Marshal(questions)
		if err != nil {
			return nil, err
		}
		err = s.redisDB.Set(ctx, cacheKey, data, time.Minute*40).Err()
		if err != nil {
			return nil, err
		}
	}
	return questions, nil
}
