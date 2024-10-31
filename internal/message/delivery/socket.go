package delivery

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	socketio "github.com/googollee/go-socket.io"

	"pineywss/config"
	"pineywss/internal/message/domain"
	"pineywss/internal/message/usecase"
)

func SocketServer(cfgRedis config.RedisConfig, cfgrabbitMQ config.RabbitMQConfig, socketPort string, messageUseCase usecase.MessageService) {
	// Create a websocket transport with a custom CheckOrigin function
	// err := messageUseCase.PublishRandomChatRoomToRabbitMQ(cfgrabbitMQ.Queues.Consumer)
	// if err != nil {
	// 	log.Println(err)
	// }

	server := socketio.NewServer(nil)
	_, _ = server.Adapter(&socketio.RedisAdapterOptions{
		Host:     cfgRedis.Host,
		Port:     strconv.Itoa(cfgRedis.Port),
		Password: cfgRedis.Pass,
	})
	log.Println("Online Rooms", server.Rooms("/"))
	messageUseCase.RunConsumer(cfgrabbitMQ.Queues.Producer, cfgrabbitMQ.Queues.Consumer, server)
	server.OnConnect("/", func(s socketio.Conn) error {
		query := s.URL().RawQuery
		values, _ := url.ParseQuery(query)
		chatId := values.Get("chatId")
		s.Leave(s.ID())
		s.Join(chatId)
		if err := messageUseCase.SetOnline(chatId); err != nil {
			log.Println(err)
		}
		go messageUseCase.GiveOldMessagesFromScyllaDB(cfgrabbitMQ.Queues.Producer, chatId, server)
		fmt.Printf("Client connected: %s\n", s.ID())
		return nil
	})
	server.OnEvent("/", "client_message", func(s socketio.Conn, msg domain.Entity) {
		// rooms := s.Rooms()
		// for r := range rooms {

		// 	messages, err := messageUseCase.SendProfitToSocket(rooms[r])
		// 	if err != nil {
		// 		log.Println(err)
		// 	}
		// 	for i := range messages {
		// 		log.Println("On clinet_message event")
		// 		s.Emit("server_message", messages[i])
		// 	}
		// }
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("Error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		query := s.URL().RawQuery
		values, _ := url.ParseQuery(query)
		chatId := values.Get("chatId")
		err := messageUseCase.DisconnectFromSocket(chatId)
		fmt.Println("Client disconnected:", reason, err)
	})

	go server.Serve()
	defer server.Close()

	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		server.ServeHTTP(w, r)
	})

	fmt.Printf("Server running on port %s...\n", socketPort)
	log.Fatal(http.ListenAndServe(socketPort, nil))
}
