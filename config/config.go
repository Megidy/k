package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	//HTTP configuration
	HTTPPort string `env:"HTTP_PORT"`

	//Worker configuration
	NumberOfWorkers int `env:"NUMBER_OF_WORKERS"`

	//JWT configuration
	JWTSecretKey string `env:"JWT_SECRET_KEY"`

	//Manager configuration
	TimeForAnswer          int `env:"TIME_FOR_ANSWER"`
	TimeForRoomLiquidation int `env:"TIME_FOR_ROOM_LIQUDATION"`

	//MySQL configuration
	MySQLConnectionString string `env:"MYSQL_CONNECTION_STRING"`
	MySQLRootPassword     string `env:"MYSQL_ROOT_PASSWORD"`
	MySQLDatabase         string `env:"MYSQL_DATABASE"`

	//Redis configuration
	RedisConnectionString         string `env:"REDIS_CONNECTION_STRING"`
	RedisTopicCacheDuration       int    `env:"REDIS_TOPIC_CACHE_DURATION"`
	RedisLeaderboardCacheDuration int    `env:"REDIS_LEADERBOARD_CACHE_DURATION"`
}

func NewConfig() *Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln("error when pasting env file : ", err)
	}
	log.Println("config :", cfg)
	return &cfg
}
