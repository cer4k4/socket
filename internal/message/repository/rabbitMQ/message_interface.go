package rabbitmq

import socketio "github.com/googollee/go-socket.io"

type MessageRabbitMQRepository interface {
	PublishMessage(queueName string, body []byte) error
	StartConsumer(prodecureQueueName, consumerqueueName string, server *socketio.Server)
}
