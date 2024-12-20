package game

import (
	"database/sql"
	"log"

	"github.com/Megidy/k/types"
	"github.com/google/uuid"
)

type store struct {
	db *sql.DB
}

func NewGameStore(db *sql.DB) *store {
	return &store{db: db}
}

func (s *store) CreateTopic(questions []types.Question) error {
	topic := questions[0].Topic
	_, err := s.db.Exec("insert into topics values(?,?,?)", topic.TopicID, uuid.NewString(), topic.Name)
	if err != nil {
		log.Println("error occured in topics : ", err)
		return err

	}
	for _, question := range questions {

		_, err = s.db.Exec("insert into questions values(?,?,?,?,?,?)", question.Id, question.Topic.TopicID, question.Type, question.ImageLink, question.Question, question.CorrectAnswer)
		if err != nil {
			log.Println("error occured in questions : ", err)
			return err
		}
		for _, answer := range question.Answers {
			_, err = s.db.Exec("insert into answers values(?,?)", question.Id, answer)
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
	rows, err := s.db.Query("select * from topics where user_id=?", userID)
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

	return topics, nil
}

func (s *store) GetQuestionsByTopicName(TopicName string) ([]types.Question, error) {
	var questions []types.Question
	row, err := s.db.Query("select * from topics where name=?", TopicName)
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

	rows, err := s.db.Query("select id,type,image_link,question,correct_answer from questions where topic_id=?", t.TopicID)
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
		rows, err := s.db.Query("select answer from answers where question_id=?", q.Id)
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
	return questions, nil
}
