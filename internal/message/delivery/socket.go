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

func MessageSocketHandler(cfgRedis config.RedisConfig, cfgrabbitMQ config.RabbitMQConfig, socketPort string, messageUseCase usecase.MessageService) {
	// Create a websocket transport with a custom CheckOrigin function

	server := socketio.NewServer(nil)
	_, _ = server.Adapter(&socketio.RedisAdapterOptions{
		Host:     cfgRedis.Host,
		Port:     strconv.Itoa(cfgRedis.Port),
		Password: cfgRedis.Pass,
	})

	server.OnConnect("/", func(s socketio.Conn) error {
		query := s.URL().RawQuery
		values, _ := url.ParseQuery(query)
		chatId := values.Get("chatId")
		s.Join(chatId)
		if err := messageUseCase.SetOnline(chatId); err != nil {
			log.Println(err)
		}
		fmt.Printf("Client connected: %s\n", s.ID())
		return nil
	})
	server.OnEvent("/", "client_message", func(s socketio.Conn, msg domain.Data) {
		rooms := s.Rooms()
		for r := range rooms {
			if messages, err := messageUseCase.ProfitToSocket(cfgrabbitMQ.Queues.Consumer, cfgrabbitMQ.Queues.Producer, rooms[r]); err != nil {
				log.Println(err)
			} else {
				for i := range messages {
					s.Emit("server_message", messages[i])
				}
			}
		}
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
