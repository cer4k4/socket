package main

import (
	"fmt"
	"log"
	"net/http"
	"pineywss/config" // Replace with your actual package path
	"pineywss/database/scylla/migrations"
	"pineywss/internal/message/delivery"
	rabbitmq "pineywss/internal/message/repository/rabbitMQ"
	"pineywss/internal/message/repository/redis"
	"pineywss/internal/message/repository/scylla"
	"pineywss/internal/message/usecase"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig("/home/aka/Templates/simple_socket/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// RabbitMQ connection string
	rabitConn, rabbitChannel := cfg.GetRabbitMQ()
	log.Println("RabbitMQ URL:", &rabitConn, rabbitChannel)

	// Get consumer queue name
	consumerQueue := cfg.RabbitMQ.Queues.Consumer
	log.Printf("Consumer Queue: %s", consumerQueue)

	// ScyllaDB connection info
	scyllaHosts := cfg.GetScyllaDBHosts()
	log.Printf("ScyllaDB Hosts: %v", scyllaHosts)
	migrations.RunMigrations(scyllaHosts)

	// Redis connection info
	redisClient := cfg.GetRedis()
	log.Printf("Redis Address: %s", redisClient)

	// Server Host & Port

	_, socketPort := cfg.GetServerAddress()
	log.Fatal(http.ListenAndServe(socketPort, nil))
	fmt.Printf("Server running on port %s...\n", socketPort)
	messageRabbitmqRepository := rabbitmq.NewMessageRabbitMQRepository(rabbitChannel)
	messageScyllaRepository := scylla.NewMessageScyllaRepository(scyllaHosts)
	messageRedisRepository := redis.NewMessageRedisRepository(redisClient)
	messageService := usecase.NewMessageService(messageRabbitmqRepository, messageScyllaRepository, messageRedisRepository)
	delivery.MessageSocketHandler(messageService)

}
