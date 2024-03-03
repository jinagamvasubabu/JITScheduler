package persistence

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	"github.com/jinagamvasubabu/JITScheduler-svc/config"
)

var redisClient *redis.Client

func NewRedisClient() *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: config.GetConfig().RedisHost,
		DB:   0,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		// logger.Info("Testing panic while Redis")
		panic(err.Error())
	}
	logger.Info("redis connected successfully")
	return redisClient
}

func GetRedisClient() *redis.Client {
	return redisClient
}

func setKeyValue(key string, value interface{}, expiration time.Duration) error {
	return GetRedisClient().SetNX(key, value, expiration).Err()
}

func GetCacheKey(serviceIdentifier string, entityIdentifier string, mapping string, key string) string {
	return strings.Join([]string{serviceIdentifier, entityIdentifier, mapping, key}, ":")
}

func GetFromRedis(key string) (*int32, error) {
	valueInt64, err := redisClient.Get(key).Int64()
	if err != nil {
		return nil, err
	}
	valueInt32 := int32(valueInt64)
	return &valueInt32, nil
}
