package usecase

import (
	"log"
	"pineywss/internal/message/domain"
	rabbitmq "pineywss/internal/message/repository/rabbitMQ"
	"pineywss/internal/message/repository/redis"
	"pineywss/internal/message/repository/scylla"
	"strconv"
)

type messageService struct {
	rabbitRepository   rabbitmq.MessageRabbitMQRepository
	scylladbRepository scylla.MessageScyllaRepository
	redisRepository    redis.MessageRedisRepository
}

func NewMessageService(rabbitRepository rabbitmq.MessageRabbitMQRepository, scylladbRepository scylla.MessageScyllaRepository, redisRepository redis.MessageRedisRepository) MessageService {
	return &messageService{rabbitRepository, scylladbRepository, redisRepository}
}

func (ms *messageService) ProfitToSocket(consumerQueue, prodecureQueue, chatroomid string) (result []domain.Data, err error) {
	// Old Messages
	oldMessage := ms.scylladbRepository.FetchMessagesFromDB(chatroomid)
	for i := range oldMessage {
		result = append(result, oldMessage[i])
		err = ms.scylladbRepository.DeleteMessage(oldMessage[i])
		if err != nil {
			return []domain.Data{}, err
		}
	}

	// New Messages
	listMsg, listData, err := ms.rabbitRepository.StartConsumer(consumerQueue)
	log.Println(len(listMsg), len(listData))
	if err != nil {
		return []domain.Data{}, err
	}
	for i := range listData {
		log.Println(listData[i].ChatId, chatroomid)
		if strconv.Itoa(int(listData[i].ChatId)) == chatroomid {
			result = append(result, listData[i])
		} else {
			err = ms.scylladbRepository.SaveMessage(listData[i])
			if err != nil {
				return []domain.Data{}, err
			}
		}
	}
	return result, err
}

func (ms *messageService) SetOnline(chatroomid string) (err error) {
	err = ms.redisRepository.SetRoomStatus(chatroomid)
	return
}

func (ms *messageService) DisconnectFromSocket(chatroomid string) (err error) {
	err = ms.redisRepository.Disconnect(chatroomid)
	return
}
