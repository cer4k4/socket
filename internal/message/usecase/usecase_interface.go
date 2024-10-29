package usecase

import "pineywss/internal/message/domain"

type MessageService interface {
	SetOnline(chatroomid string) error
	DisconnectFromSocket(chatroomid string) error
	PublishRandomChatRoomToRabbitMQ(prodecureQueue string) (err error)
	SendProfitToSocket(chatroomid string) ([]domain.Data, error)
	GiveMessagesFromRabbit(queueName string)
}
