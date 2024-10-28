package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"pineywss/internal/message/domain"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type messageRabbit struct {
	channel *amqp091.Channel
}

func NewMessageRabbitMQRepository(channel *amqp091.Channel) MessageRabbitMQRepository {
	return &messageRabbit{channel}
}

func (m *messageRabbit) PublishMessage(queueName string, body []byte) error {
	_, err := m.channel.QueueDeclare(
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

	err = m.channel.Publish(
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

func (m *messageRabbit) StartConsumer(queueName string) ([]domain.Message, []domain.Data, error) {

	var listMsg []domain.Message
	var listData []domain.Data
	maxMessages := 10

	if queueName == "" {
		return nil, nil, fmt.Errorf("CONSUMER_QUEUE environment variable not set")
	}

	// Declare queue
	_, err := m.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare queue %s: %v", queueName, err)
	}

	// Start consuming
	msgs, err := m.channel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register a consumer: %v", err)
	}

	timeoutChan := time.After(time.Second)
	messageCount := 0
	msgChan := make(chan domain.Message)
	dataChan := make(chan domain.Data)
	errorChan := make(chan error)
	done := make(chan bool)

	// Process messages in goroutine
	go func() {
		for d := range msgs {
			var msg domain.Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				errorChan <- fmt.Errorf("error parsing message: %v", err)
				continue
			}

			var data domain.Data
			if err := json.Unmarshal([]byte(msg.Data), &data); err != nil {
				errorChan <- fmt.Errorf("error parsing data: %v", err)
				continue
			}

			msgChan <- msg
			dataChan <- data

			messageCount++
			if messageCount >= maxMessages {
				done <- true
				return
			}
		}
	}()

	// Collect results
	for {
		select {
		case <-timeoutChan:
			log.Printf("Timeout reached while consuming messages. Processed %d messages", messageCount)
			return listMsg, listData, nil

		case msg := <-msgChan:
			listMsg = append(listMsg, msg)

		case data := <-dataChan:
			listData = append(listData, data)

		case err := <-errorChan:
			log.Printf("Error while processing message: %v", err)

		case <-done:
			log.Printf("Processed %d messages successfully", messageCount)
			return listMsg, listData, nil
		}
	}
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
