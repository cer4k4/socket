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

func (ms *messageScylla) SaveMessage(data domain.Data) error {
	query := `INSERT INTO messages (chat_id, profit, failed_update_attempts) VALUES (?, ?, ?)`
	return ms.session.Query(query, data.ChatId, data.Profit, data.FailedUpdateAttempts).Exec()
}

func (ms *messageScylla) DeleteMessage(data domain.Data) error {
	query := `DELETE FROM messages WHERE chat_id = ?`
	return ms.session.Query(query, data.ChatId).Exec()
}

func (ms *messageScylla) FetchMessagesFromDB(chatId string) []domain.Data {
	var messages []domain.Data
	chatIdInt, _ := strconv.ParseInt(chatId, 10, 64)

	iter := ms.session.Query("SELECT profit, failed_update_attempts FROM messages WHERE chat_id = ?", chatIdInt).Iter()
	var profit, failedUpdateAttempts int
	for iter.Scan(&profit, &failedUpdateAttempts) {
		messages = append(messages, domain.Data{
			ChatId:               chatIdInt,
			Profit:               profit,
			FailedUpdateAttempts: failedUpdateAttempts,
		})
	}
	if err := iter.Close(); err != nil {
		log.Printf("Error closing ScyllaDB iterator for user %s: %v", chatId, err)
	}
	return messages
}
