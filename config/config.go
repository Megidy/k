package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// HTTP configuration
	HTTPPort string `env:"HTTP_PORT" envDefault:":8080"`

	// Worker configuration
	NumberOfWorkers int `env:"NUMBER_OF_WORKERS" envDefault:"100"`

	// JWT configuration
	JWTSecretKey string `env:"JWT_SECRET_KEY" envDefault:"f0834ujf90r3fhrju38fc9r3hf093hr3"`

	// Manager configuration
	TimeForAnswer          int `env:"TIME_FOR_ANSWER" envDefault:"20"`
	TimeForRoomLiquidation int `env:"TIME_FOR_ROOM_LIQUDATION" envDefault:"360"`

	// MySQL configuration
	MySQLConnectionString string `env:"MYSQL_CONNECTION_STRING" envDefault:"root:password@tcp(mysql:3306)/k"`
	MySQLRootPassword     string `env:"MYSQL_ROOT_PASSWORD" envDefault:"password"`
	MySQLDatabase         string `env:"MYSQL_DATABASE" envDefault:"k"`

	// Redis configuration
	RedisConnectionString         string `env:"REDIS_CONNECTION_STRING" envDefault:"redis:6379"`
	RedisTopicCacheDuration       int    `env:"REDIS_TOPIC_CACHE_DURATION" envDefault:"600"`
	RedisLeaderboardCacheDuration int    `env:"REDIS_LEADERBOARD_CACHE_DURATION" envDefault:"60000"`
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
