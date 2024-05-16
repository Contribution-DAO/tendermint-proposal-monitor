package proposals

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Proposal struct {
	ProposalID string `json:"proposal_id"`
	Status     string `json:"status"`
	Content    struct {
		Title string `json:"title"`
	} `json:"content"`
	VotingEndTime string `json:"voting_end_time"`
}

func mockProposals() []Proposal {
	return []Proposal{
		{
			ProposalID: "300",
			Status:     "PROPOSAL_STATUS_VOTING_PERIOD",
			Content: struct {
				Title string `json:"title"`
			}{
				Title: "Proposal 1 Title",
			},
			VotingEndTime: "2024-05-20T00:00:00.725539835Z",
		},
	}
}

func Fetch(apiEndpoint string, useMock bool) ([]Proposal, error) {
	if useMock {
		return mockProposals(), nil
	}

	resp, err := http.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch proposals: %s", resp.Status)
	}

	var result struct {
		Proposals []Proposal `json:"proposals"`
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Proposals, nil
}

func GetAlertedProposals(filename string) (map[string]bool, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]bool), nil
		}
		return nil, err
	}

	var alertedProposals map[string]bool
	err = json.Unmarshal(data, &alertedProposals)
	if err != nil {
		return nil, err
	}

	return alertedProposals, nil
}

func SaveAlertedProposals(filename string, alertedProposals map[string]bool) error {
	data, err := json.Marshal(alertedProposals)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}
