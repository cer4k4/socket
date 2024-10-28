package usecase

import "pineywss/internal/message/domain"

type MessageService interface {
	ProfitToSocket(consumerQueue, prodecureQueue, chatroomid string) ([]domain.Data, error)
	SetOnline(chatroomid string) error
	DisconnectFromSocket(chatroomid string) error
}
