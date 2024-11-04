package rabbitmq

type MessageRabbitMQRepository interface {
	PublishMessage(queueName string, body []byte) error
	StartConsumer(prodecureQueueName, consumerqueueName string)
}
