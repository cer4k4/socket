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
	isOnline, _ := mr.redisClient.Get(context.Background(), "user:"+chatRoomId+":online").Result()
	// if err != nil {
	// 	log.Printf("Error geting online status for user %s in Redis: %v", chatRoomId, err)
	// }
	return isOnline, nil
}

func (mr *messageRedis) SetRoomStatus(chatroomid string) error {
	err := mr.redisClient.Set(context.Background(), "user:"+chatroomid+":online", "true", 0).Err()
	if err != nil {
		log.Printf("Error setting online status for user %s in Redis: %v", chatroomid, err)
		return err
	}
	return redis.Nil
}

func (mr *messageRedis) Disconnect(chatroomid string) error {
	err := mr.redisClient.Del(context.Background(), "user:"+chatroomid+":online").Err()
	if err != nil {
		log.Printf("Error removing online status for user %s in Redis: %v", chatroomid, err)
	}
	return err
}
