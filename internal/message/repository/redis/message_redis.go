package redis

import (
	"context"
	"log"

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

func (mr *messageRedis) SetRoomStatus(chatroomid string) error {
	err := mr.redisClient.Set(context.Background(), "user:"+chatroomid+":online", "true", 0).Err()
	if err != nil {
		log.Printf("Error setting online status for user %s in Redis: %v", chatroomid, err)
		return err
	}
	return nil
}

func (mr *messageRedis) Disconnect(chatroomid string) error {
	err := mr.redisClient.Del(context.Background(), "user:"+chatroomid+":online").Err()
	if err != nil {
		log.Printf("Error removing online status for user %s in Redis: %v", chatroomid, err)
	}
	return err
}
