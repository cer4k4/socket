package usecase

import (
	rabbitmq "pineywss/internal/message/repository/rabbitMQ"
	"pineywss/internal/message/repository/redis"
	"pineywss/internal/message/repository/scylla"
)

type messageService struct {
	rabbitRepository   rabbitmq.MessageRabbitMQRepository
	scylladbRepository scylla.MessageScyllaRepository
	redisRepository    redis.MessageRedisRepository
}

func NewMessageService(rabbitChannel rabbitmq.MessageRabbitMQRepository, scylladbRepository scylla.MessageScyllaRepository, redisRepository redis.MessageRedisRepository) MessageService {
	return &messageService{rabbitChannel, scylladbRepository, redisRepository}
}

func (ms *messageService) NewProfitToSocket() {

}
