package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Server   ServerConfig `yaml:"server"`
	Database DBConfig     `yaml:"database"`
	Secret   string       `yaml:"secret"`
	AppId    int64        `yaml:"appId"`
	//VkApiKey       string       `yaml:"vkApiKey"`
	//VkGroupId      string       `yaml:"vkGroupId"`
	//VkGroupAlbumId string       `yaml:"vkGroupAlbumId"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
