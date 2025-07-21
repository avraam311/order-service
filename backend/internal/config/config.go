package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   Server   `yaml:"server"`
	Logger   Logger   `yaml:"logger"`
	Database Database `yaml:"database"`
	Kafka    Kafka    `yaml:"kafka"`
	Cache    Cache    `yaml:"cache"`
}

type Server struct {
	HTTPPort string `yaml:"httpPort"`
}

type Logger struct {
	Env         string `yaml:"env"`
	LogFilePath string `yaml:"logFilePath"`
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string `yaml:"sslmode"`
}

type Kafka struct {
	GroupID string   `yaml:"groupID"`
	Topic   string   `yaml:"topic"`
	Brokers []string `yaml:"brokers"`
}

type Cache struct {
	DefaultExpiration time.Duration `yaml:"defaultExpiration"`
	CleanupInterval   time.Duration `yaml:"cleanupInterval"`
	PreloadLimit      int           `yaml:"preloadLimit"`
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func MustLoad() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("fatal error config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	cfg.Database.Host = os.Getenv("DB_HOST")
	cfg.Database.Port = os.Getenv("DB_PORT")
	cfg.Database.User = os.Getenv("DB_USER")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	cfg.Database.Name = os.Getenv("DB_NAME")

	return &cfg
}
