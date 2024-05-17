package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ProposalDetailDomain       string                 `yaml:"proposal_detail_domain"`
	VotingAlertBehaviorNearing string                 `yaml:"voting_alert_behavior_nearing"`
	Discord                    DiscordConfig          `yaml:"discord"`
	Chains                     map[string]ChainConfig `yaml:"chains"`
}

type DiscordConfig struct {
	Enabled bool   `yaml:"enabled"`
	Webhook string `yaml:"webhook"`
}

type ChainConfig struct {
	ChainID          string      `yaml:"chain_id"`
	ValidatorAddress string      `yaml:"validator_address"`
	APIVersion       string      `yaml:"api_version"`
	APIEndpoint      string      `yaml:"api_endpoint"`
	ExplorerURL      string      `yaml:"explorer_url"`
	Alerts           AlertConfig `yaml:"alerts"`
}

type AlertConfig struct {
	Discord struct {
		Enabled bool   `yaml:"enabled"`
		Webhook string `yaml:"webhook"`
	} `yaml:"discord"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
