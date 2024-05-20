package monitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/notifiers"
	"tendermint_proposal_monitor/proposals"
	"tendermint_proposal_monitor/utils"
)

func SendDiscordAlert(cfg *config.Configurations, chain config.ChainConfig, chainName string, proposal proposals.Proposal, globalDiscordNotifier *notifiers.DiscordNotifier, alertType string) error {
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
