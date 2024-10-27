package scylla

import (
	"pineywss/internal/message/domain"

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
