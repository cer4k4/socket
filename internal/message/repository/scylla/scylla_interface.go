package scylla

import (
	"pineywss/internal/message/domain"
)

type MessageScyllaRepository interface {
	SaveMessage(data domain.Data) error
	DeleteMessage(data domain.Data) error
}
