package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/Megidy/k/types"
	"github.com/redis/go-redis/v9"
)

var pictures = []string{
	"https://i.pinimg.com/736x/b0/93/c5/b093c578f3c99b2525194db73cf12e01.jpg",
	"https://i.pinimg.com/736x/b2/7f/31/b27f31e8ddc96d54536e1f162948272a.jpg",
	"https://i.pinimg.com/736x/f8/b2/20/f8b220ee6d7f3c12b9c1ba3f202a5813.jpg",
	"https://i.pinimg.com/736x/6e/89/fa/6e89faaffc9ff42df7167c60abf6775c.jpg",
	"https://i.pinimg.com/736x/f0/07/f5/f007f57c6092bf3ca8189756de467365.jpg",
}

type store struct {
	db      *sql.DB
	redisDB *redis.Client
}

func NewStore(db *sql.DB, redisDB *redis.Client) *store {
	return &store{db: db, redisDB: redisDB}
}

func (s *store) GetUserById(id string) (*types.User, error) {
	row, err := s.db.Query("SELECT * FROM users WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	var user types.User
	for row.Next() {
		err = row.Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.Description, &user.ProfilePicture)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}
func (s *store) CreateUser(user *types.User) error {
	user.ProfilePicture = pictures[rand.Intn(5)]

	_, err := s.db.Exec("INSERT INTO users(id,username,email,password,profile_picture) VALUES(?,?,?,?,?)", user.ID, user.UserName, user.Email, user.Password, user.ProfilePicture)
	if err != nil {
		return err
	}
	return nil
}
func (s *store) UserExists(user *types.User) (bool, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE email=? OR username=?", user.Email, user.UserName)
	if err != nil {
		return false, err
	}
	for !rows.Next() {
		return false, nil
	}
	return true, nil
}

func (s *store) GetUserByEmail(email string) (*types.User, error) {
	rows, err := s.db.Query("SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, err
	}
	var user types.User
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.Description, &user.ProfilePicture)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *store) UpdateDescription(userID, description string) error {
	_, err := s.db.Exec("UPDATE users SET description=? WHERE id=?", description, userID)

	if err != nil {
		return err
	}

	return nil
}

func (s *store) CacheUserGameScore(username, score, place, topicName string) error {
	cacheKey := "user:" + username + ":leaderboard"
	ctx := context.Background()

	val, err := s.redisDB.Get(ctx, cacheKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	var leaderBoard []types.UserLeaderBoard
	if val != "" {
		if err := json.Unmarshal([]byte(val), &leaderBoard); err != nil {
			return err
		}
	}

	leaderBoard = append(leaderBoard, types.UserLeaderBoard{Score: score, Place: place, TopicName: topicName})

	jsonData, err := json.Marshal(leaderBoard)
	if err != nil {
		return err
	}

	return s.redisDB.Set(ctx, cacheKey, jsonData, time.Second*60).Err()
}

func (s *store) GetUserGameScore(username string) ([]types.UserLeaderBoard, bool, error) {
	cacheKey := "user:" + username + ":leaderboard"
	ctx := context.Background()

	val, err := s.redisDB.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	var leaderBoard []types.UserLeaderBoard
	if err := json.Unmarshal([]byte(val), &leaderBoard); err != nil {
		return nil, false, err
	}

	return leaderBoard, true, nil
}
