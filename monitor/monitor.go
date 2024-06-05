package monitor

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/notifiers"
	"tendermint_proposal_monitor/proposals"
	"tendermint_proposal_monitor/services"
)

type Handler struct {
	Services *services.NewServices
}

func NewHandler(services *services.NewServices) *Handler {
	return &Handler{
		Services: services,
	}
}

// Define constants for alert types and file names
const (
	AlertTypeNewProposal   = "📝 New proposal on"
	AlertTypeVotingNearing = "🕒 Voting period is nearing its end"
)

// Define constant for voting alert behavior
const (
	VotingAlertBehaviorOnlyIfNotVoted = "only_if_not_voted"
)

func (h *Handler) Run(cfg *config.Configurations, useMock bool) error {
	ctx := context.Background()
	lastChecked, alertedProposals, votingEndAlertedProposals, err := h.Services.FirestoreHandler.InitState()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("error init state: %v", err)
	}

	globalDiscordNotifier := &notifiers.DiscordNotifier{WebhookURL: cfg.Discord.Webhook}

	log.Printf("Checking for new proposals...")

	for chainName, chain := range cfg.Chains {
		propList, err := fetchProposals(chain, useMock, chainName)
		if err != nil {
			continue
		}

		for _, proposal := range propList {
			if shouldSkipProposal(proposal) {
				continue
			}

			proposalID, err := strconv.Atoi(proposal.ProposalID)
			if err != nil {
				log.Printf("Invalid proposal ID: %v", err)
				continue
			}

			// Check if the proposal is new and alert if it hasn't been alerted yet
			if proposalID > lastChecked[chainName] {
				err = SendDiscordAlert(cfg, chain, chainName, proposal, globalDiscordNotifier, AlertTypeNewProposal)
				if err != nil {
					log.Printf("Error sending alert for new proposal: %v", err)
					continue
				}
				lastChecked[chainName] = proposalID
				if alertedProposals[chainName] == nil {
					alertedProposals[chainName] = make(map[string]bool)
				}
				alertedProposals[chainName][proposal.ProposalID] = true
				err = h.Services.FirestoreHandler.SaveLastCheckedProposalIDs(ctx, proposals.CollectionNameLastChecked, lastChecked)
				if err != nil {
					log.Printf("Error saving last checked proposal ID: %v", err)
				}

				err = h.Services.FirestoreHandler.SaveAlertedProposals(ctx, proposals.CollectionNameAlertedProposals, alertedProposals)
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
						err = SendDiscordAlert(cfg, chain, chainName, proposal, globalDiscordNotifier, AlertTypeVotingNearing)
						if err != nil {
							log.Printf("Error sending alert for voting nearing end: %v", err)
							continue
						}
						if votingEndAlertedProposals[chainName] == nil {
							votingEndAlertedProposals[chainName] = make(map[string]bool)
						}
						votingEndAlertedProposals[chainName][proposal.ProposalID] = true

						err = h.Services.FirestoreHandler.SaveAlertedProposals(ctx, proposals.CollectionNameVotingEndAlerted, votingEndAlertedProposals)
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

func fetchProposals(chain config.ChainConfig, useMock bool, chainName string) ([]proposals.Proposal, error) {
	propList, err := proposals.Fetch(chain, chain.APIVersion, useMock)
	if err != nil {
		log.Printf("Error fetching proposals for chain %s: %v", chainName, err)
	}
	return propList, err
}

func shouldSkipProposal(proposal proposals.Proposal) bool {
	if statusValue, exists := proposals.ProposalStatusValue[proposal.Status]; exists &&
		(statusValue == proposals.ProposalStatusValue["PROPOSAL_STATUS_PASSED"] ||
			statusValue == proposals.ProposalStatusValue["PROPOSAL_STATUS_REJECTED"] ||
			statusValue == proposals.ProposalStatusValue["PROPOSAL_STATUS_FAILED"]) {
		return true
	}
	return false
}
