package redis

type MessageRedisRepository interface {
	GetRoomStatus(chatRoomId string) (string, error)
}
