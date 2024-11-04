package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"pineywss/internal/message/domain"
	"pineywss/internal/message/repository/redis"
	"pineywss/internal/message/repository/scylla"
	"strconv"

	"github.com/ambelovsky/gosf"

	"github.com/rabbitmq/amqp091-go"
)

type messageRabbit struct {
	rabbitConn         *amqp091.Connection
	redisRepository    redis.MessageRedisRepository
	scylladbRepository scylla.MessageScyllaRepository
}

func NewMessageRabbitMQRepository(rabbitConn *amqp091.Connection, redisRepository redis.MessageRedisRepository, scylladbRepository scylla.MessageScyllaRepository) MessageRabbitMQRepository {
	return &messageRabbit{rabbitConn, redisRepository, scylladbRepository}
}

func (m *messageRabbit) PublishMessage(queueName string, body []byte) error {
	ch, err := m.rabbitConn.Channel()
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare queue %s: %v", queueName, err)
		return err
	}

	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish message to queue %s: %v", queueName, err)
	}
	return err
}

func (m *messageRabbit) StartConsumer(prodecureQueueName, consumerqueueName string) {
	ch, err := m.rabbitConn.Channel()
	if consumerqueueName == "" {
		fmt.Printf("CONSUMER_QUEUE environment variable not set")
	}

	// Declare queue
	queue, err := ch.QueueDeclare(
		consumerqueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("failed to declare queue %s: %v", consumerqueueName, err)
	}

	messageCount := queue.Messages

	err = ch.Qos(
		int(messageCount), // prefetch count - only get existing messages
		0,                 // prefetch size
		false,             // global
	)
	if err != nil {
		fmt.Printf("failed to set QoS: %v", err)
	}

	// Start consuming
	msgs, err := ch.Consume(
		consumerqueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("failed to register a consumer: %v", err)
	}
	var count int
	var chMessage = make(chan domain.Message, 1000)
	var chData = make(chan domain.Data, 1000)
	go func() {
		for d := range msgs {
			var msg domain.Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}

			var data domain.Data
			if err := json.Unmarshal([]byte(msg.Data), &data); err != nil {
				log.Printf("Error parsing data: %v", err)
				continue
			}
			log.Printf("Counsum From (%v) Queue, Value Of Message %v \n", consumerqueueName, msg)
			if err != nil {
				log.Println(err)
			}
			count++
			chMessage <- msg
			chData <- data
		}
	}()
	go func() {
		for {
			select {
			case msg := <-chMessage:
				data := <-chData
				stchatroom := strconv.Itoa(int(data.ChatId))
				statusRoom, _ := m.redisRepository.GetRoomStatus(stchatroom)
				if statusRoom == "true" {
					gosf.Broadcast(stchatroom, "response", &gosf.Message{Body: gosf.StructToMap(msg)})
					//server.BroadcastToRoom("/", stchatroom, "server_message", msg)
					log.Printf("Online: Send To Room %v \n", data)
					dataByte, _ := json.Marshal(msg)
					m.PublishMessage(prodecureQueueName, dataByte)
					log.Printf("Send To (%v) Queue \n", prodecureQueueName)
				} else {
					log.Printf("Offline SaveMessage %v %v On Scylla \n", msg, data)
					if err := m.scylladbRepository.SaveMessage(msg, data); err != nil {
						log.Println("You Have Error In Save To ScyllaDB", err)
					}
				}

			}
		}
	}()
}

/*	chatIdStr := strconv.FormatInt(data.ChatId, 10)

	isOnline, err := redisClient.Get(context.Background(), "user:"+chatIdStr+":online").Result()
	if err != nil || isOnline != "true" {
		err = models.SaveMessage(scyllaSession, data)
		if err != nil {
			log.Printf("Error saving message to database: %v", err)
		}
		continue
	}

	socketServer.SendMessage(chatIdStr, "MinerProfit", msg)

	err = models.DeleteMessage(scyllaSession, data)
	if err != nil {
		log.Printf("Error deleting message from database: %v", err)
	}


	PublishMessage(channel, os.Getenv("PRODUCER_QUEUE"), d.Body)
*/

//New Code

// dataChan := make(chan domain.Data, messageCount)
// done := make(chan bool)
// processedCount := 0
// go func() {
// 	//		rooms := server.Rooms("/")
// 	var data domain.Data
// 	messageCount := 0
// 	for d := range msgs {

// 		if err := json.Unmarshal(d.Body, &data); err != nil {
// 			log.Printf("Error unmarshaling message: %v", err)
// 			d.Nack(false, true)
// 			continue
// 		}
// 		server.BroadcastToRoom("/", "6627322", "client_message")
// 		dataChan <- data
// 		d.Ack(false) // Acknowledge message
// 		processedCount++
// 		messageCount++
// 		log.Printf("Processed message %d: %+v", messageCount, data)
// 		if processedCount >= int(messageCount) {
// 			defer close(dataChan)
// 			return
// 		}
// 	}
// }()
// // Consumer goroutine

// go func() {
// 	for data := range dataChan {
// 		mu.Lock()
// 		listData = append(listData, data)
// 		mu.Unlock()
// 	}
// 	done <- true
// }()
// // go func() {
// // 	for {
// // 		select {
// // 		case data := <-dataChan:
// // 			mu.Lock()
// // 			listData = append(listData, data)
// // 			mu.Unlock()
// // 		case <-timeout:
// // 			done <- true
// // 			return
// // 		}
// // 	}
// // }()

// // Wait for completion or timeout
// select {
// case <-done:
// 	log.Printf("Completed processing. Total messages: %d", len(listData))
// 	return listMsg, listData, nil
// case <-time.After(1 * time.Minute): // Overall timeout
// 	log.Println("Consumer timed out")
// 	return listMsg, listData, nil
// }
