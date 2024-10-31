package scylla

import (
	"pineywss/internal/message/domain"
)

type MessageScyllaRepository interface {
	SaveMessage(msg domain.Message, data domain.Data) error
	DeleteMessage(chatroomid int64) error
	FetchMessagesFromDB(chatId string) []domain.Message
}
