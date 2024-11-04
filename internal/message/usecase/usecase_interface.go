package usecase

type MessageService interface {
	SetOnline(chatroomid string) error
	DisconnectFromSocket(chatroomid string) error
	PublishRandomChatRoomToRabbitMQ(prodecureQueue string) (err error)
	//SendProfitToSocket(chatroomid string) ([]domain.Entity, error)
	RunConsumer(prodecureQueueName, consumerQueueName string)
	GiveOldMessagesFromScyllaDB(prodecureQueueName, chatId string)
}
