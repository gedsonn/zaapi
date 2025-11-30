package config

import (
	"os"
	"sync"

	"github.com/creasty/defaults"
	"github.com/goccy/go-yaml"
)


type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Enable  bool   `yml:"enable"`
	Debug   bool   `yaml:"debug"`
	Maneger bool   `yaml:"maneger"`
	Swagger bool   `yaml:"swagger"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type WebhookConfig struct {
	Global  string `yaml:"global"`
	Local   string `yaml:"local"`
	Enabled bool   `yaml:"enabled"`

	Timeout int `yaml:"timeout"`
	Retries int `yaml:"retries"`
}

type WhatsConfig struct {
	Version string `yaml:"version"`
}

type Configuration struct {
	Name     string         `yaml:"name"`
	Server   ServerConfig   `yaml:"server"`
	Token    string         `yaml:"token"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Webhook  WebhookConfig  `yaml:"webhook"`
	Whatsapp WhatsConfig    `yaml:"whatsapp"`
}

var (
	mu      sync.RWMutex
	_config *Configuration
)

func NewAtPath(path string) (*Configuration, error) {
	c := Configuration{
		Name:  "Zaapi",
		Token: "token",

		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			Enable:  true,
			Debug:   false,
			Maneger: false,
			Swagger: false,
		},

		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "zaapi",
			Password: "zaapi",
			Name:     "zaapi",
		},

		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},

		Webhook: WebhookConfig{
			Global:  "http://localhost:8080",
			Local:   "http://localhost:8080",
			Enabled: false,
			Timeout: 5,
			Retries: 3,
		},

		Whatsapp: WhatsConfig{
			Version: "latest",
		},
	}

	if err := defaults.Set(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func Set(c *Configuration) {
	mu.Lock()
	defer mu.Unlock()
	token := c.Token
	if token == "" {
		token = "zaapi-default-token"
	}
	_config = c
	_config.Token = token
}

func Get() *Configuration {
	mu.RLock()
	c := *_config
	mu.RUnlock()
	return &c
}

func FromFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	c, err := NewAtPath(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(b, c); err != nil {
		return err
	}

	Set(c)
	return nil
}

func Default() *Configuration {
	c, _ := NewAtPath("")
	return c
}

func Save(path string) error {
	mu.RLock()
	c := *_config
	mu.RUnlock()

	data, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}