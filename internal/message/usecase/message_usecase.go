package usecase

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
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
	chanData           chan domain.Data
}

func NewMessageService(rabbitRepository rabbitmq.MessageRabbitMQRepository, scylladbRepository scylla.MessageScyllaRepository, redisRepository redis.MessageRedisRepository, chanData chan domain.Data) MessageService {
	return &messageService{rabbitRepository, scylladbRepository, redisRepository, chanData}
}

func (ms *messageService) SendProfitToSocket(chatroomid string) (result []domain.Data, err error) {
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
	select {
	case data := <-ms.chanData:
		intChatId, _ := strconv.Atoi(chatroomid)
		if int64(intChatId) == data.ChatId {
			log.Println("New Messafges", data)
			result = append(result, data)
			return result, nil
		}
	default:
		return result, nil
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

func (ms *messageService) PublishRandomChatRoomToRabbitMQ(prodecureQueue string) (err error) {
	for i := 0; i < 10; i++ {
		//chatroomid, _ := GenerateInRange(1000000, 10000000)
		var data domain.Data
		data.ChatId = int64(6627322)
		data.Profit = 12
		bdata, _ := json.Marshal(data)
		err = ms.rabbitRepository.PublishMessage(prodecureQueue, bdata)
	}
	return err
}

func (ms *messageService) GiveMessagesFromRabbit(queueName string) {
	_, listData, err := ms.rabbitRepository.StartConsumer(queueName)
	log.Println("Messages Count", len(listData))
	if err != nil {
		log.Printf("After Give RabbitMQ, %v \n", err)
	}
	for l := range listData {
		strchatid := strconv.Itoa(int(listData[l].ChatId))
		status, err := ms.redisRepository.GetRoomStatus(strchatid)
		if err != nil {
			log.Printf("Status Redis, %v \n", err)
		}
		if status != "true" {
			err := ms.scylladbRepository.SaveMessage(listData[l])
			if err != nil {
				log.Printf("SaveMessage To ScyllaDB, %v \n", err)
			}
		} else {
			log.Printf("Send to Channel From GiveMessagesFromRabbit, %v \n", listData[l])
			ms.chanData <- listData[l]
		}
	}
}

func GenerateInRange(min, max int) (int, error) {
	if min > max {
		min, max = max, min // Swap if min is greater than max
	}

	diff := max - min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 0, err
	}

	return int(n.Int64()) + min, nil
}
