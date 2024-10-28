package redis

type MessageRedisRepository interface {
	GetRoomStatus(chatRoomId string) (string, error)
	SetRoomStatus(chatroomid string) error
	Disconnect(chatroomid string) error
}
