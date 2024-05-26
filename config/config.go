package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Configurations struct {
	ProposalDetailDomain       string                 `yaml:"proposal_detail_domain"`
	VotingAlertBehaviorNearing string                 `yaml:"voting_alert_behavior_nearing"`
	Discord                    DiscordConfig          `yaml:"discord"`
	Chains                     map[string]ChainConfig `yaml:"chains"`
	Storage                    Storage                `yaml:"storage"`
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

type Storage struct {
	CredentialsPath string `yaml:"credentials_path"`
	ProjectID       string `yaml:"project_id"`
	DatabaseID      string `yaml:"database_id"`
	CollectionName  string `yaml:"table_name"`
}

func LoadConfig(filename string) (*Configurations, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Configurations
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
