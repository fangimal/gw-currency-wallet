package config

import (
	"gw-currency-wallet/pkg/logging"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogIsDebug    *bool         `yaml:"log_is_debug" env-default:"true"`
	ExchangerAddr string        `yaml:"exchanger_addr" env-default:"50052"`
	JWTSecret     string        `yaml:"jwt_secret" env-default:"your-very-long-secret-key-here"`
	HTTPPort      string        `yaml:"http_port" env-default:"8080"`
	KafkaBroker   string        `yaml:"kafka_broker" env-default:"localhost:9092"`
	KafkaTopic    string        `yaml:"kafka_topic" env-default:"large-transfers"`
	Storage       StorageConfig `yaml:"storage"`
}

type StorageConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     string `yaml:"port" env-default:"5432"`
	User     string `yaml:"username" env-default:"wallet_user"`
	Password string `yaml:"password" env-default:"123"`
	Name     string `yaml:"database" env-default:"wallet_db"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	logger := logging.GetLogger()
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Infof("Error reading config: %v", help)
		}
	})
	return instance
}
