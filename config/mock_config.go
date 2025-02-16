package config

func NewMockConfig(timeForAnswer, timeForRoomLiquidation int) *Config {
	return &Config{
		HTTPPort:                      ":8080",
		NumberOfWorkers:               100,
		JWTSecretKey:                  "mocked-secret-key",
		TimeForAnswer:                 timeForAnswer,
		TimeForRoomLiquidation:        timeForRoomLiquidation,
		MySQLConnectionString:         "mockuser:mockpassword@tcp(mock-mysql:3306)/mockdb",
		MySQLRootPassword:             "mockpassword",
		MySQLDatabase:                 "mockdb",
		RedisConnectionString:         "mock-redis:6379",
		RedisTopicCacheDuration:       48,
		RedisLeaderboardCacheDuration: 48,
	}
}
