package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppStage string `env:"APP_STAGE" envDefault:"dev"`
	Port     int    `env:"PORT" envDefault:"3002"`
	Host     string `env:"HOST" envDefault:"localhost"`

	MongoURI      string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDatabase string `env:"MONGO_DATABASE" envDefault:"analytics_service"`

	RabbitMQURL      string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672"`
	RabbitMQExchange string `env:"RABBITMQ_ORDER_EVENTS_EXCHANGE" envDefault:"order-service.events"`
	RabbitMQQueue    string `env:"RABBITMQ_QUEUE" envDefault:"analytics-service.order-events"`
	RabbitMQDLX      string `env:"RABBITMQ_DLX" envDefault:"analytics-service.order-events.dlx"`
	RabbitMQDLQ      string `env:"RABBITMQ_DLQ" envDefault:"analytics-service.order-events.dead"`
	RabbitMQPrefetch int    `env:"RABBITMQ_PREFETCH" envDefault:"10"`

	CoreEventsExchange string `env:"RABBITMQ_CORE_EVENTS_EXCHANGE" envDefault:"core-service.events"`
	CoreEventsQueue    string `env:"RABBITMQ_CORE_EVENTS_QUEUE" envDefault:"analytics-service.core-events"`

	AccessSecret string `env:"ACCESS_SECRET,required"`

	CoreServiceURL    string `env:"CORE_SERVICE_URL" envDefault:"http://localhost:3000"`
	CoreServiceAPIKey string `env:"CORE_SERVICE_API_KEY" envDefault:"internal-api-key"`

	RBACCacheTTLSec   int `env:"RBAC_CACHE_TTL_SEC" envDefault:"300"`
	EventDedupTTLDays int `env:"EVENT_DEDUP_TTL_DAYS" envDefault:"7"`
}

func Load() (*Config, error) {
	stage := os.Getenv("APP_STAGE")
	if stage == "" {
		stage = "dev"
	}

	var envFile string
	switch stage {
	case "dev":
		envFile = ".env.dev"
	case "test":
		envFile = ".env.test"
	default:
		envFile = ".env"
	}

	root, _ := os.Getwd()
	path := filepath.Join(root, envFile)
	_ = godotenv.Load(path)

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	return &cfg, nil
}
