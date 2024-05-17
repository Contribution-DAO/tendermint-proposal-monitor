package proposals

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Proposal represents a governance proposal with common fields for both v1 and v1beta1 endpoints
type Proposal struct {
	ProposalID      string `json:"proposal_id"`
	Status          string `json:"status"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	VotingStartTime string `json:"voting_start_time"`
	VotingEndTime   string `json:"voting_end_time"`
}

// ProposalV1 represents the structure for v1 API responses
type ProposalV1 struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Messages []struct {
		Content struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"content"`
	} `json:"messages"`
	VotingStartTime string `json:"voting_start_time"`
	VotingEndTime   string `json:"voting_end_time"`
}

// ProposalV1Beta1 represents the structure for v1beta1 API responses
type ProposalV1Beta1 struct {
	ProposalID string `json:"proposal_id"`
	Status     string `json:"status"`
	Content    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"content"`
	VotingStartTime string `json:"voting_start_time"`
	VotingEndTime   string `json:"voting_end_time"`
}

func mockProposals() []Proposal {
	return []Proposal{
		{
			ProposalID:      "44",
			Status:          "PROPOSAL_STATUS_VOTING_PERIOD",
			Title:           "Governance Community Spend Guardrails",
			Description:     "Introduction: As a community, it is important to ensure that we have a way to control community.",
			VotingStartTime: "2024-05-15T00:00:00.725539835Z",
			VotingEndTime:   "2024-05-18T00:00:00.725539835Z",
		},
	}
}

func Fetch(apiEndpoint, sdkVersion string, useMock bool) ([]Proposal, error) {
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch sdkVersion {
	case "v1":
		var result struct {
			Proposals []ProposalV1 `json:"proposals"`
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		return mapProposalsV1(result.Proposals), nil

	case "v1beta1":
		var result struct {
			Proposals []ProposalV1Beta1 `json:"proposals"`
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		return mapProposalsV1Beta1(result.Proposals), nil

	default:
		return nil, fmt.Errorf("unsupported sdk version: %s", sdkVersion)
	}
}

func mapProposalsV1(proposals []ProposalV1) []Proposal {
	var mapped []Proposal
	for _, p := range proposals {
		if len(p.Messages) > 0 {
			title := "No Title"
			description := "No Description"

			if len(p.Messages) > 0 && p.Messages[0].Content.Title != "" {
				title = p.Messages[0].Content.Title
			}
			if len(p.Messages) > 0 && p.Messages[0].Content.Description != "" {
				description = p.Messages[0].Content.Description
			}

			mapped = append(mapped, Proposal{
				ProposalID:      p.ID,
				Status:          p.Status,
				Title:           title,
				Description:     description,
				VotingStartTime: p.VotingStartTime,
				VotingEndTime:   p.VotingEndTime,
			})
		} else {
			mapped = append(mapped, Proposal{
				ProposalID:      p.ID,
				Status:          p.Status,
				Title:           "No Title",
				Description:     "No Description",
				VotingStartTime: p.VotingStartTime,
				VotingEndTime:   p.VotingEndTime,
			})
		}
	}
	return mapped
}

func mapProposalsV1Beta1(proposals []ProposalV1Beta1) []Proposal {
	var mapped []Proposal
	for _, p := range proposals {
		mapped = append(mapped, Proposal{
			ProposalID:      p.ProposalID,
			Status:          p.Status,
			Title:           p.Content.Title,
			Description:     p.Content.Description,
			VotingStartTime: p.VotingStartTime,
			VotingEndTime:   p.VotingEndTime,
		})
	}
	return mapped
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
