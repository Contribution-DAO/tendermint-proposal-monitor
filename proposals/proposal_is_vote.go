package proposals

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tendermint_proposal_monitor/config"
)

type VoteOption struct {
	Option string `json:"option"`
	Weight string `json:"weight"`
}

type VoteResponseV1 struct {
	Vote struct {
		ProposalID string       `json:"proposal_id"`
		Voter      string       `json:"voter"`
		Options    []VoteOption `json:"options"`
		Metadata   string       `json:"metadata"`
	} `json:"vote"`
}

type VoteResponseV1Beta1 struct {
	Vote struct {
		ProposalID string       `json:"proposal_id"`
		Voter      string       `json:"voter"`
		Options    []VoteOption `json:"options"`
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

	switch sdkVersion {
	case "v1":
		var voteResponse VoteResponseV1
		err = json.NewDecoder(resp.Body).Decode(&voteResponse)
		if err != nil {
			return false, fmt.Errorf("error decoding vote response for proposal %s: %w", proposalID, err)
		}
		if voteResponse.Vote.Voter == validatorAddress {
			return true, nil
		}
	case "v1beta1":
		var voteResponse VoteResponseV1Beta1
		err = json.NewDecoder(resp.Body).Decode(&voteResponse)
		if err != nil {
			return false, fmt.Errorf("error decoding vote response for proposal %s: %w", proposalID, err)
		}
		if voteResponse.Vote.Voter == validatorAddress {
			return true, nil
		}
	default:
		return false, fmt.Errorf("unsupported sdk version: %s", sdkVersion)
	}

	return false, nil
}
