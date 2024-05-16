package monitor

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/notifiers"
	"tendermint_proposal_monitor/proposals"
)

func Run(cfg *config.Config, useMock bool) error {
	lastChecked, err := proposals.GetLastCheckedProposalIDs("data/last_checked_proposals.json")
	if err != nil {
		log.Printf("Error loading last checked proposal IDs, defaulting to empty: %v", err)
		lastChecked = make(map[string]int)
	}

	alertedProposals, err := proposals.GetAlertedProposals("data/alerted_proposals.json")
	if err != nil {
		log.Printf("Error loading alerted proposals, defaulting to empty: %v", err)
		alertedProposals = make(map[string]bool)
	}

	votingEndAlertedProposals, err := proposals.GetAlertedProposals("data/voting_end_alerted_proposals.json")
	if err != nil {
		log.Printf("Error loading voting end alerted proposals, defaulting to empty: %v", err)
		votingEndAlertedProposals = make(map[string]bool)
	}

	globalDiscordNotifier := &notifiers.DiscordNotifier{WebhookURL: cfg.Discord.Webhook}

	for {
		log.Printf("Checking for new proposals every %d seconds...\n", cfg.CheckInterval)

		for chainName, chain := range cfg.Chains {
			propList, err := proposals.Fetch(chain.Alerts.APIEndpoint, useMock)
			if err != nil {
				log.Printf("Error fetching proposals for chain %s: %v", chainName, err)
				continue
			}

			for _, proposal := range propList {
				if statusValue, exists := proposals.ProposalStatusValue[proposal.Status]; exists &&
					(statusValue == proposals.ProposalStatusValue["PROPOSAL_STATUS_PASSED"] ||
						statusValue == proposals.ProposalStatusValue["PROPOSAL_STATUS_REJECTED"]) {
					continue
				}

				proposalID, err := strconv.Atoi(proposal.ProposalID)
				if err != nil {
					log.Printf("Invalid proposal ID: %v", err)
					continue
				}

				// Check if the proposal is new and alert if it hasn't been alerted yet
				if proposalID > lastChecked[chainName] {
					sendDiscordAlert(cfg, chain, chainName, proposal, globalDiscordNotifier, "New proposal detected on chain")
					lastChecked[chainName] = proposalID
					alertedProposals[proposal.ProposalID] = true
					err = proposals.SaveLastCheckedProposalIDs("data/last_checked_proposals.json", lastChecked)
					if err != nil {
						log.Printf("Error saving last checked proposal ID: %v", err)
					}
					err = proposals.SaveAlertedProposals("data/alerted_proposals.json", alertedProposals)
					if err != nil {
						log.Printf("Error saving alerted proposals: %v", err)
					}
				}

				// Check if the proposal is nearing its voting end time and if it has not been alerted yet
				if proposal.Status == "PROPOSAL_STATUS_VOTING_PERIOD" {
					votingEndTime, err := time.Parse(time.RFC3339, proposal.VotingEndTime)
					if err != nil {
						log.Printf("Error parsing voting end time: %v", err)
						continue
					}
					currentTime := time.Now()
					if !votingEndAlertedProposals[proposal.ProposalID] && votingEndTime.Sub(currentTime) <= 24*time.Hour {
						sendDiscordAlert(cfg, chain, chainName, proposal, globalDiscordNotifier, "Voting period is nearing its end")
						votingEndAlertedProposals[proposal.ProposalID] = true
						err = proposals.SaveAlertedProposals("data/voting_end_alerted_proposals.json", votingEndAlertedProposals)
						if err != nil {
							log.Printf("Error saving voting end alerted proposals: %v", err)
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(cfg.CheckInterval) * time.Second)
	}
}

func sendDiscordAlert(cfg *config.Config, chain config.ChainConfig, chainName string, proposal proposals.Proposal, globalDiscordNotifier *notifiers.DiscordNotifier, alertType string) {
	var discordNotifier *notifiers.DiscordNotifier
	if chain.Alerts.Discord.Enabled && chain.Alerts.Discord.Webhook != "" {
		discordNotifier = &notifiers.DiscordNotifier{WebhookURL: chain.Alerts.Discord.Webhook}
	} else if cfg.Discord.Enabled && cfg.Discord.Webhook != "" {
		discordNotifier = globalDiscordNotifier
	} else {
		log.Printf("No valid Discord webhook URL available for chain %s", chainName)
		return
	}

	// Send an alert to Discord if enabled
	message := fmt.Sprintf("%s: %s\nTitle: %s\nStatus: %s\nVoting End Time: %s",
		alertType, chainName, proposal.Content.Title, proposal.Status, proposal.VotingEndTime)
	err := discordNotifier.SendAlert(message)
	if err != nil {
		log.Printf("Error sending Discord alert: %v", err)
	}
}
