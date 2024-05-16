package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type DiscordConfig struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
}

type HealthcheckConfig struct {
	Enabled  bool   `yaml:"enabled"`
	PingURL  string `yaml:"ping_url"`
	PingRate int    `yaml:"ping_rate"`
}

type ChainAlertConfig struct {
	APIEndpoint string        `yaml:"api_endpoint"`
	Discord     DiscordConfig `yaml:"discord"`
}

type ChainConfig struct {
	ChainID string           `yaml:"chain_id"`
	Alerts  ChainAlertConfig `yaml:"alerts"`
}

type Config struct {
	CheckInterval int                    `yaml:"check_interval"`
	Discord       DiscordConfig          `yaml:"discord"`
	Healthcheck   HealthcheckConfig      `yaml:"healthcheck"`
	Chains        map[string]ChainConfig `yaml:"chains"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
