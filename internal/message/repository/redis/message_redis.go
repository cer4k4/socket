package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type messageRedis struct {
	redisClient *redis.Client
}

func NewMessageRedisRepository(redisClient *redis.Client) MessageRedisRepository {
	return &messageRedis{redisClient}
}

func (mr *messageRedis) GetRoomStatus(chatRoomId string) (string, error) {
	isOnline, err := mr.redisClient.Get(context.Background(), "user:"+chatRoomId+":online").Result()
	return isOnline, err
}
