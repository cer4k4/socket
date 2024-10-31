package scylla

import (
	"log"
	"pineywss/internal/message/domain"
	"strconv"

	"github.com/gocql/gocql"
)

type messageScylla struct {
	session *gocql.Session
}

func NewMessageScyllaRepository(session *gocql.Session) MessageScyllaRepository {
	return &messageScylla{session}
}

func (ms *messageScylla) SaveMessage(msg domain.Message, data domain.Data) error {
	//query := `INSERT INTO messages (id, chat_id, profit, failed_update_attempts) VALUES (uuid(),?, ?, ?)`
	//return ms.session.Query(query, data.Data.ChatId, data.Data.Profit, data.Data.FailedUpdateAttempts).Exec()
	query := `INSERT INTO messages (id, entity_type, operation, data, chat_id, profit, failed_update_attempts) VALUES (uuid(), ?, ?, ?, ?, ?, ?);`
	err := ms.session.Query(query, msg.EntityType, msg.Operation, msg.Data, data.ChatId, data.Profit, data.FailedUpdateAttempts).Exec()
	log.Println("db layer", err)
	return err
}

func (ms *messageScylla) DeleteMessage(chatroomid int64) error {
	query := `DELETE FROM messages WHERE chat_id = ?`
	return ms.session.Query(query, chatroomid).Exec()
}

func (ms *messageScylla) FetchMessagesFromDB(chatId string) []domain.Message {
	var messages []domain.Message
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 64)

	iter := ms.session.Query("SELECT entity_type, operation, data, profit, failed_update_attempts FROM messages WHERE chat_id = ? ALLOW FILTERING", chatIdInt).Iter()
	var entityType, operation, profit, failedUpdateAttempts int
	var data string
	for iter.Scan(&entityType, &operation, &data, &profit, &failedUpdateAttempts) {
		messages = append(messages, domain.Message{
			EntityType: entityType,
			Operation:  operation,
			Data:       data,
		})
	}
	if err := iter.Close(); err != nil {
		log.Printf("Error closing ScyllaDB iterator for user %s: %v", chatId, err)
	}
	return messages
}
