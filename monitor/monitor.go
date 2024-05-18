package monitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/notifiers"
	"tendermint_proposal_monitor/proposals"
	"tendermint_proposal_monitor/utils"
)

// Define constants for alert types and file names
const (
	AlertTypeNewProposal   = "📝 New proposal on"
	AlertTypeVotingNearing = "🕒 Voting period is nearing its end"
	FileLastChecked        = "data/last_checked_proposals.json"
	FileAlertedProposals   = "data/alerted_proposals.json"
	FileVotingEndAlerted   = "data/voting_end_alerted_proposals.json"
)

// Define constant for voting alert behavior
const (
	VotingAlertBehaviorOnlyIfNotVoted = "only_if_not_voted"
)

func Run(cfg *config.Config, useMock bool) error {
	lastChecked, err := proposals.GetLastCheckedProposalIDs(FileLastChecked)
	if err != nil {
		log.Printf("Error loading last checked proposal IDs, defaulting to empty: %v", err)
		lastChecked = make(map[string]int)
	}

	alertedProposals, err := proposals.GetAlertedProposals(FileAlertedProposals)
	if err != nil {
		log.Printf("Error loading alerted proposals, defaulting to empty: %v", err)
		alertedProposals = make(map[string]map[string]bool)
	}

	votingEndAlertedProposals, err := proposals.GetAlertedProposals(FileVotingEndAlerted)
	if err != nil {
		log.Printf("Error loading voting end alerted proposals, defaulting to empty: %v", err)
		votingEndAlertedProposals = make(map[string]map[string]bool)
	}

	globalDiscordNotifier := &notifiers.DiscordNotifier{WebhookURL: cfg.Discord.Webhook}

	log.Printf("Checking for new proposals...")

	for chainName, chain := range cfg.Chains {
		propList, err := proposals.Fetch(chain, chain.APIVersion, useMock)
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
				err = sendDiscordAlert(cfg, chain, chainName, proposal, globalDiscordNotifier, AlertTypeNewProposal)
				if err != nil {
					log.Printf("Error sending alert for new proposal: %v", err)
					continue
				}
				lastChecked[chainName] = proposalID
				if alertedProposals[chainName] == nil {
					alertedProposals[chainName] = make(map[string]bool)
				}
				alertedProposals[chainName][proposal.ProposalID] = true
				err = proposals.SaveLastCheckedProposalIDs(FileLastChecked, lastChecked)
				if err != nil {
					log.Printf("Error saving last checked proposal ID: %v", err)
				}
				err = proposals.SaveAlertedProposals(FileAlertedProposals, alertedProposals)
				if err != nil {
					log.Printf("Error saving alerted proposals: %v", err)
				}
			}

			// Check if the proposal is nearing its voting end time and if it has not been alerted yet
			if proposal.Status == proposals.ProposalStatusName[1] {
				votingEndTime, err := time.Parse(time.RFC3339, proposal.VotingEndTime)
				if err != nil {
					log.Printf("Error parsing voting end time: %v", err)
					continue
				}
				currentTime := time.Now()
				if !votingEndAlertedProposals[chainName][proposal.ProposalID] && votingEndTime.Sub(currentTime) <= 24*time.Hour {
					shouldSendAlert := true

					if cfg.VotingAlertBehaviorNearing == VotingAlertBehaviorOnlyIfNotVoted {
						voted, err := proposals.CheckValidatorVoted(chain, proposal.ProposalID, chain.ValidatorAddress, chain.APIVersion)
						if err != nil {
							log.Printf("%v", err)
							continue
						}

						if voted {
							shouldSendAlert = false
						}
					}

					if shouldSendAlert {
						err = sendDiscordAlert(cfg, chain, chainName, proposal, globalDiscordNotifier, AlertTypeVotingNearing)
						if err != nil {
							log.Printf("Error sending alert for voting nearing end: %v", err)
							continue
						}
						if votingEndAlertedProposals[chainName] == nil {
							votingEndAlertedProposals[chainName] = make(map[string]bool)
						}
						votingEndAlertedProposals[chainName][proposal.ProposalID] = true
						err = proposals.SaveAlertedProposals(FileVotingEndAlerted, votingEndAlertedProposals)
						if err != nil {
							log.Printf("Error saving voting end alerted proposals: %v", err)
						}
					}
				}
			}
		}
	}

	return nil
}

func sendDiscordAlert(cfg *config.Config, chain config.ChainConfig, chainName string, proposal proposals.Proposal, globalDiscordNotifier *notifiers.DiscordNotifier, alertType string) error {
	var discordNotifier *notifiers.DiscordNotifier

	if chain.Alerts.Discord.Enabled && chain.Alerts.Discord.Webhook != "" {
		discordNotifier = &notifiers.DiscordNotifier{WebhookURL: chain.Alerts.Discord.Webhook}
	} else if chain.Alerts.Discord.Enabled && (cfg.Discord.Enabled && cfg.Discord.Webhook != "") {
		discordNotifier = globalDiscordNotifier
	} else {
		log.Printf("No valid Discord webhook URL available for chain %s", chainName)
		return fmt.Errorf("no valid Discord webhook URL available for chain %s", chainName)
	}

	proposalDetail := utils.GenerateProposalDetailURL(cfg.ProposalDetailDomain, chainName, proposal.ProposalID)
	if chain.ExplorerURL != "" {
		proposalDetail = fmt.Sprintf("%s/%s", chain.ExplorerURL, proposal.ProposalID)
	}

	endTime, err := time.Parse(time.RFC3339Nano, proposal.VotingEndTime)
	if err != nil {
		return fmt.Errorf("error parsing voting end time: %v", err)
	}
	timeLeft := utils.FormatTimeLeft(endTime)

	description := proposal.Description
	if len(description) > 120 {
		description = description[:117] + "..."
	}

	votingStartTime, err := time.Parse(time.RFC3339, proposal.VotingStartTime)
	if err != nil {
		return fmt.Errorf("error parsing VotingStartTime: %v", err)
	}

	formattedVotingStartTime := votingStartTime.Format("2006-01-02 15:04")

	messageContent := fmt.Sprintf("**%s %s**: %s\n\n**Proposal title:** %s\n\n**Short text description:** %s\n\n**Vote start:** %s\n\n**Time left: %s**\n\n**Read full proposal details:**\n%s",
		alertType, chainName, proposal.ProposalID, proposal.Title, description, formattedVotingStartTime, timeLeft, proposalDetail)
	embed := notifiers.DiscordEmbed{
		Color:       notifiers.MessageBoxColor,
		Description: messageContent,
	}

	discordMessage := notifiers.DiscordMessage{
		Content: "",
		TTS:     false,
		Embeds:  []notifiers.DiscordEmbed{embed},
	}

	payload, err := json.Marshal(discordMessage)
	if err != nil {
		return fmt.Errorf("error marshalling Discord message: %v", err)
	}

	resp, err := discordNotifier.SendPayload(payload)
	if err != nil {
		return fmt.Errorf("error sending Discord alert: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error sending Discord alert, response status: %d, response body: %s", resp.StatusCode, string(body))
	}

	return nil
}
