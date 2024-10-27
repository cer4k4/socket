package delivery

import (
	"fmt"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"

	"pineywss/internal/message/domain"
	"pineywss/internal/message/usecase"
)

func MessageSocketHandler(messageUseCase usecase.MessageService) {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		fmt.Printf("Client connected: %s\n", s.ID())
		return nil
	})

	server.OnEvent("/", "client_message", func(s socketio.Conn, msg domain.Message) {
		fmt.Printf("Received from client: %s\n", msg.Data)
		queueName := "profitEvent"
		if queueName == "" {
			log.Fatal("CONSUMER_QUEUE environment variable not set")
		}
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("Error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("Client disconnected:", reason)
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

}
