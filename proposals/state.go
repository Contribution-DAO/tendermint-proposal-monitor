package proposals

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type LastCheckedProposals struct {
	Proposals map[string]int `json:"proposals"`
}

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
