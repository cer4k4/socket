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

	"github.com/ambelovsky/gosf"
)

type messageService struct {
	rabbitRepository   rabbitmq.MessageRabbitMQRepository
	scylladbRepository scylla.MessageScyllaRepository
	redisRepository    redis.MessageRedisRepository
}

func NewMessageService(rabbitRepository rabbitmq.MessageRabbitMQRepository, scylladbRepository scylla.MessageScyllaRepository, redisRepository redis.MessageRedisRepository) MessageService {
	return &messageService{rabbitRepository, scylladbRepository, redisRepository}
}

// func (ms *messageService) SendProfitToSocket(chatroomid string) (result []domain.Message, err error) {
// 	// Old Messages
// 	oldMessage := ms.scylladbRepository.FetchMessagesFromDB(chatroomid)
// 	for i := range oldMessage {
// 		result = append(result, oldMessage[i])
// 		err = ms.scylladbRepository.DeleteMessage(oldMessage[i])
// 		if err != nil {
// 			return []domain.Message{}, err
// 		}
// 	}

// 	// New Messages
// 	select {
// 	case data := <-ms.chanData:
// 		intChatId, _ := strconv.Atoi(chatroomid)
// 		if int64(intChatId) == data.Data.ChatId {
// 			log.Println("New Messafges", data)
// 			result = append(result, data)
// 			return result, nil
// 		}
// 	default:
// 		return result, nil
// 	}

// 	return result, err

// }

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

func (ms *messageService) RunConsumer(prodecureQueueName, consumerQueueName string) {
	ms.rabbitRepository.StartConsumer(prodecureQueueName, consumerQueueName)
}

func (ms *messageService) GiveOldMessagesFromScyllaDB(prodecureQueueName, chatId string) {
	oldmessageList := ms.scylladbRepository.FetchMessagesFromDB(chatId)
	for l := range oldmessageList {
		gosf.Broadcast(chatId, "response", &gosf.Message{Body: gosf.StructToMap(oldmessageList[l])})
		intChatId, _ := strconv.Atoi(chatId)
		err := ms.scylladbRepository.DeleteMessage(int64(intChatId))
		if err != nil {
			log.Println("err delete db", err)
		}
		byteData, _ := json.Marshal(oldmessageList[l])
		ms.rabbitRepository.PublishMessage(prodecureQueueName, byteData)
		log.Println("after Send to Socket offline messages send to dbEventQueue", oldmessageList[l])
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
