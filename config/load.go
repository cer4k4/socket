package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gocql/gocql"
	"github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	ScyllaDB ScyllaDBConfig `yaml:"scylladb"`
	Server   ServerConfig   `yaml:"server"`
	Redis    RedisConfig    `yaml:"redis"`
}

type RabbitMQConfig struct {
	User   string      `yaml:"user"`
	Pass   string      `yaml:"pass"`
	Host   string      `yaml:"host"`
	Port   int         `yaml:"port"`
	Queues QueueConfig `yaml:"queues"`
}

type QueueConfig struct {
	Consumer string `yaml:"consumer"`
	Producer string `yaml:"producer"`
}

type ScyllaDBConfig struct {
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Keyspace string `yaml:"keyspace"`
}

type ServerConfig struct {
	Host       string `yaml:"host"`
	SocketPort int    `yaml:"socket_port"`
	HTTPPort   int    `yaml:"http_port"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Pass string `yaml:"pass"`
}

// LoadConfig reads the YAML configuration file and returns a Config struct
func LoadConfig(filename string) (*Config, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(buf, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return config, nil
}

// GetRabbitMQURL returns RabbitMQ connection
func (c *Config) GetRabbitMQ() (*amqp091.Connection, *amqp091.Channel) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		c.RabbitMQ.User,
		c.RabbitMQ.Pass,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
	)
	conn, err := amqp091.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	return conn, channel

}

// GetScyllaDBHosts returns the ScyllaDB
func (c *Config) GetScyllaDBHosts() *gocql.Session {

	cluster := gocql.NewCluster(fmt.Sprintf("%s:%d", c.ScyllaDB.Host, c.ScyllaDB.Port))
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: c.ScyllaDB.User,
		Password: c.ScyllaDB.Pass,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Unable to connect to ScyllaDB: %v", err)
	}
	defer session.Close()

	// Ensure the keyspace exists
	err = ensureKeyspace(session, c.ScyllaDB.Keyspace)
	if err != nil {
		log.Fatalf("Failed to ensure keyspace exists: %v", err)
	}

	// Reconnect with the keyspace specified
	cluster.Keyspace = c.ScyllaDB.Keyspace
	sessionWithKeyspace, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Unable to connect to ScyllaDB with keyspace: %v", err)
	}

	return sessionWithKeyspace
}

func ensureKeyspace(session *gocql.Session, keyspace string) error {
	// Check if keyspace exists
	meta, err := session.KeyspaceMetadata(keyspace)
	if err == nil && meta != nil {
		// Keyspace exists
		return nil
	}

	// Keyspace does not exist, create it
	cql := fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'}`, keyspace)
	return session.Query(cql).Exec()
}

// GetRedisAddr returns the Redis Client
func (c *Config) GetRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port),
		Password: c.Redis.Pass,
		DB:       0,
	})

	// Test connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return client
}

func (c *Config) GetServerAddress() (string, string) {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.HTTPPort), fmt.Sprintf("%s:%d", c.Server.Host, c.Server.SocketPort)
}
