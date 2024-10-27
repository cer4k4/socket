package rabbitmq

import (
	"encoding/json"
	"log"
	"pineywss/internal/message/domain"

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
	if queueName == "" {
		log.Fatal("CONSUMER_QUEUE environment variable not set")
	}

	_, err := m.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue %s: %v", queueName, err)
		return []domain.Message{}, []domain.Data{}, err
	}

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
		log.Fatalf("Failed to register a consumer: %v", err)
		return []domain.Message{}, []domain.Data{}, err
	}

	go func() {
		for d := range msgs {
			var msg domain.Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}
			listMsg = append(listMsg, msg)
			var data domain.Data
			if err := json.Unmarshal([]byte(msg.Data), &data); err != nil {
				log.Printf("Error parsing data: %v", err)
				continue
			}

			listData = append(listData, data)

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
		}
	}()
	log.Printf("Consuming messages from queue %s...", queueName)
	return listMsg, listData, err
}
