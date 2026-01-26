package config

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogIsDebug *bool `yaml:"log_is_debug" env-default:"true"`
	GRPC       struct {
		Port string `yaml:"port" env-default:"50051"`
	} `yaml:"grpc"`
	Storage   StorageConfig `yaml:"storage"`
	JWTSecret string        `yaml:"jwt_secret" env-default:"your-very-long-secret-key-here"`
	HTTPPort  string        `yaml:"http_port" env-default:"8080"`
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
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			fmt.Printf("Error reading config: %v", help)
		}
	})
	return instance
}
