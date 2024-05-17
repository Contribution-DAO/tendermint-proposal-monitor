package proposals

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tendermint_proposal_monitor/config"
)

type VoteResponse struct {
	Vote struct {
		ProposalID string `json:"proposal_id"`
		Voter      string `json:"voter"`
		Options    []struct {
			Option string `json:"option"`
			Weight string `json:"weight"`
		} `json:"options"`
		Metadata string `json:"metadata"`
	} `json:"vote"`
}

func CheckValidatorVoted(chain config.ChainConfig, proposalID string, validatorAddress string, sdkVersion string) (bool, error) {
	voteCheckURL := fmt.Sprintf("%s/cosmos/gov/%s/proposals/%s/votes/%s", chain.Alerts.APIEndpoint, sdkVersion, proposalID, validatorAddress)
	resp, err := http.Get(voteCheckURL)
	if err != nil {
		return false, fmt.Errorf("error fetching vote status for proposal %s: %w", proposalID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var voteResponse VoteResponse
	err = json.NewDecoder(resp.Body).Decode(&voteResponse)
	if err != nil {
		return false, fmt.Errorf("error decoding vote response for proposal %s: %w", proposalID, err)
	}

	if voteResponse.Vote.Voter == validatorAddress {
		return true, nil
	}

	return false, nil
}
