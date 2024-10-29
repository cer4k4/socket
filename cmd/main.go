package main

import (
	"log"
	"pineywss/config" // Replace with your actual package path
	"pineywss/database/scylla/migrations"
	"pineywss/internal/message/delivery"
	"pineywss/internal/message/domain"
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

	// ScyllaDB connection info
	scyllaHosts := cfg.GetScyllaDBHosts()
	log.Printf("ScyllaDB Hosts: %v", scyllaHosts)
	migrations.RunMigrations(scyllaHosts)

	// Redis connection info
	redisClient := cfg.GetRedis()
	log.Printf("Redis Address: %s", redisClient)

	// Server Host & Port

	_, socketPort := cfg.GetServerAddress()
	chanData := make(chan domain.Data, 100)
	messageRabbitmqRepository := rabbitmq.NewMessageRabbitMQRepository(rabbitChannel)
	messageScyllaRepository := scylla.NewMessageScyllaRepository(scyllaHosts)
	messageRedisRepository := redis.NewMessageRedisRepository(redisClient)
	messageService := usecase.NewMessageService(messageRabbitmqRepository, messageScyllaRepository, messageRedisRepository, chanData)
	delivery.SocketServer(cfg.Redis, cfg.RabbitMQ, socketPort, messageService)

}
