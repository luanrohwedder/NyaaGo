package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Feeder      FeederConfig      `yaml:"feed"`
		Qbittorrent QbittorrentConfig `yaml:"qbittorrent"`
	}

	FeederConfig struct {
		Name    string `yaml:"name"`
		BaseURL string `yaml:"baseURL"`
	}

	QbittorrentConfig struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Port     string `yaml:"port"`
	}
)

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) UpdateConfig() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile("config.yaml", data, 0600)
}
