package usecase

import (
	socketio "github.com/googollee/go-socket.io"
)

type MessageService interface {
	SetOnline(chatroomid string) error
	DisconnectFromSocket(chatroomid string) error
	PublishRandomChatRoomToRabbitMQ(prodecureQueue string) (err error)
	//SendProfitToSocket(chatroomid string) ([]domain.Entity, error)
	RunConsumer(prodecureQueueName, consumerQueueName string, server *socketio.Server)
	GiveOldMessagesFromScyllaDB(prodecureQueueName, chatId string, server *socketio.Server)
}
