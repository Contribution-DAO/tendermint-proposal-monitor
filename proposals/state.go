package proposals

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type LastCheckedProposals struct {
	Proposals map[string]int `json:"proposals"`
}

const (
	FileLastChecked      = "data/last_checked_proposals.json"
	FileAlertedProposals = "data/alerted_proposals.json"
	FileVotingEndAlerted = "data/voting_end_alerted_proposals.json"
)

func GetLastCheckedProposalIDs(filename string) (map[string]int, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]int), nil
		}
		return nil, err
	}

	var lastChecked LastCheckedProposals
	err = json.Unmarshal(data, &lastChecked)
	if err != nil {
		return nil, err
	}

	return lastChecked.Proposals, nil
}

func SaveLastCheckedProposalIDs(filename string, lastChecked map[string]int) error {
	data, err := json.Marshal(LastCheckedProposals{Proposals: lastChecked})
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func GetAlertedProposals(filename string) (map[string]map[string]bool, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]map[string]bool), nil
		}
		return nil, err
	}

	var alertedProposals map[string]map[string]bool
	err = json.Unmarshal(data, &alertedProposals)
	if err != nil {
		return nil, err
	}

	return alertedProposals, nil
}

func SaveAlertedProposals(filename string, alertedProposals map[string]map[string]bool) error {
	data, err := json.Marshal(alertedProposals)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func InitState() (map[string]int, map[string]map[string]bool, map[string]map[string]bool, error) {
	lastChecked, err := GetLastCheckedProposalIDs(FileLastChecked)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading last checked proposal IDs, defaulting to empty: %v", err)
	}

	alertedProposals, err := GetAlertedProposals(FileAlertedProposals)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading alerted proposals, defaulting to empty: %v", err)
	}

	votingEndAlertedProposals, err := GetAlertedProposals(FileVotingEndAlerted)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading voting end alerted proposals, defaulting to empty: %v", err)
	}

	return lastChecked, alertedProposals, votingEndAlertedProposals, nil
}
