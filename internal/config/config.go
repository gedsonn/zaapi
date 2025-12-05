package config

import (
	"errors"
	"os"
	"sync"

	"github.com/goccy/go-yaml"
)

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Enable  bool   `yaml:"enable"`
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
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

type WhatsConfig struct {
	Version string `yaml:"version"`
}

type MediaConfig struct {
	Path string `yaml:"path"`
}

type Configuration struct {
	Name     string         `yaml:"name"`
	Token    string         `yaml:"token"`
	Secret   string         `yaml:"secret"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Webhook  WebhookConfig  `yaml:"webhook"`
	Whatsapp WhatsConfig    `yaml:"whatsapp"`
	Media    MediaConfig    `yaml:"media"`
}

var (
	mu      sync.RWMutex
	_config *Configuration
)

// Load loads the configuration from a file.
// If the file does not exist, it creates one with default values.
func Load(path string) (*Configuration, error) {

	// -------- READ FILE --------
	file, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		// File does not exist → create default config
		cfg := DefaultConfig()
		Set(cfg)

		if err := Save(path); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	if err != nil {
		return nil, err
	}

	// -------- PARSE YAML --------
	var data Configuration
	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	Set(&data)
	return &data, nil
}

// DefaultConfig returns a new config with predefined values.
func DefaultConfig() *Configuration {
	return &Configuration{
		Name:   "Zaapi",
		Token:  "token",
		Secret: "1234",

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

		Media: MediaConfig{
			Path: "assets",
		},
	}
}

// Set sets the config safely in memory.
func Set(c *Configuration) {
	mu.Lock()
	defer mu.Unlock()

	if c.Token == "" {
		c.Token = "zaapi-default-token"
	}

	_config = c
}

// Get returns a copy of the current configuration.
func Get() *Configuration {
	mu.RLock()
	defer mu.RUnlock()

	if _config == nil {
		return DefaultConfig() // prevenção de nil pointer
	}

	cpy := *_config
	return &cpy
}

// Save writes the config to disk.
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
