package rabbitmq

import "pineywss/internal/message/domain"

type MessageRabbitMQRepository interface {
	PublishMessage(queueName string, body []byte) error
	StartConsumer(queueName string) ([]domain.Message, []domain.Data, error)
}
