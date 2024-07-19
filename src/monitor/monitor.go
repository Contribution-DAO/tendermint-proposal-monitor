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

type ProcessProposalContext struct {
	Cfg                       *config.Configurations
	Chain                     config.ChainConfig
	ChainName                 string
	GlobalDiscordNotifier     *notifiers.DiscordNotifier
	LastChecked               map[string]int
	AlertedProposals          map[string]map[string]bool
	VotingEndAlertedProposals map[string]map[string]bool
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
	lastChecked, alertedProposals, votingEndAlertedProposals, err := h.Services.FirestoreHandler.InitState()
	if err != nil {
		log.Printf("error init state: %v", err)
		return fmt.Errorf("error init state: %v", err)
	}

	globalDiscordNotifier := &notifiers.DiscordNotifier{WebhookURL: cfg.Discord.Webhook}

	proposalCtx := &ProcessProposalContext{
		Cfg:                       cfg,
		GlobalDiscordNotifier:     globalDiscordNotifier,
		LastChecked:               lastChecked,
		AlertedProposals:          alertedProposals,
		VotingEndAlertedProposals: votingEndAlertedProposals,
	}

	log.Printf("Checking for new proposals...")

	for chainName, chain := range cfg.Chains {
		propList, err := fetchProposals(chain, useMock, chainName)
		if err != nil {
			continue
		}

		proposalCtx.Chain = chain
		proposalCtx.ChainName = chainName

		err = h.processProposals(propList, proposalCtx)
		if err != nil {
			log.Printf("Error processing proposals for chain %s: %v", chainName, err)
		}
	}

	return nil
}

func (h *Handler) processProposals(propList []proposals.Proposal, pctx *ProcessProposalContext) error {
	ctx := context.Background()
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
		err = h.checkAndSendNewProposalAlert(ctx, pctx, proposal, proposalID)
		if err != nil {
			log.Printf("Error checking new proposal alert: %v", err)
			continue
		}

		// Check if the proposal is nearing its voting end time and if it has not been alerted yet
		err = h.checkAndSendVotingNearingAlert(ctx, pctx, proposal)
		if err != nil {
			log.Printf("Error checking voting nearing alert: %v", err)
			continue
		}
	}
	return nil
}

func (h *Handler) checkAndSendNewProposalAlert(ctx context.Context, pctx *ProcessProposalContext, proposal proposals.Proposal, proposalID int) error {
	if proposalID > pctx.LastChecked[pctx.ChainName] {
		err := SendDiscordAlert(pctx.Cfg, pctx.Chain, pctx.ChainName, proposal, pctx.GlobalDiscordNotifier, AlertTypeNewProposal)
		if err != nil {
			return fmt.Errorf("error sending alert for new proposal: %v", err)
		}
		pctx.LastChecked[pctx.ChainName] = proposalID
		if pctx.AlertedProposals[pctx.ChainName] == nil {
			pctx.AlertedProposals[pctx.ChainName] = make(map[string]bool)
		}
		pctx.AlertedProposals[pctx.ChainName][proposal.ProposalID] = true

		err = h.saveState(ctx, pctx)
		if err != nil {
			return fmt.Errorf("error saving state: %v", err)
		}
	}
	return nil
}

func (h *Handler) checkAndSendVotingNearingAlert(ctx context.Context, pctx *ProcessProposalContext, proposal proposals.Proposal) error {
	if proposal.Status != proposals.ProposalStatusName[1] {
		return nil
	}

	votingEndTime, err := time.Parse(time.RFC3339, proposal.VotingEndTime)
	if err != nil {
		return fmt.Errorf("error parsing voting end time: %v", err)
	}

	currentTime := time.Now()
	if !pctx.VotingEndAlertedProposals[pctx.ChainName][proposal.ProposalID] && votingEndTime.Sub(currentTime) <= 24*time.Hour {
		shouldSendAlert, err := shouldSendVotingNearingAlert(pctx.Cfg, pctx.Chain, proposal)
		if err != nil {
			return err
		}

		if shouldSendAlert {
			err = sendVotingNearingAlert(ctx, h, pctx, proposal)
			if err != nil {
				return fmt.Errorf("error sending alert for voting nearing end: %v", err)
			}
		}
	}

	return nil
}

func (h *Handler) saveState(ctx context.Context, pctx *ProcessProposalContext) error {
	err := h.Services.FirestoreHandler.SaveLastCheckedProposalIDs(ctx, proposals.CollectionNameLastChecked, pctx.LastChecked)
	if err != nil {
		return fmt.Errorf("error saving last checked proposal ID: %v", err)
	}

	err = h.Services.FirestoreHandler.SaveAlertedProposals(ctx, proposals.CollectionNameAlertedProposals, pctx.AlertedProposals)
	if err != nil {
		return fmt.Errorf("error saving alerted proposals: %v", err)
	}

	return nil
}

func shouldSendVotingNearingAlert(cfg *config.Configurations, chain config.ChainConfig, proposal proposals.Proposal) (bool, error) {
	if cfg.VotingAlertBehaviorNearing == VotingAlertBehaviorOnlyIfNotVoted {
		voted, err := proposals.CheckValidatorVoted(chain, proposal.ProposalID, chain.ValidatorAddress, chain.APIVersion)
		if err != nil {
			return false, err
		}
		if voted {
			return false, nil
		}
	}
	return true, nil
}

func sendVotingNearingAlert(ctx context.Context, h *Handler, pctx *ProcessProposalContext, proposal proposals.Proposal) error {
	err := SendDiscordAlert(pctx.Cfg, pctx.Chain, pctx.ChainName, proposal, pctx.GlobalDiscordNotifier, AlertTypeVotingNearing)
	if err != nil {
		return err
	}
	if pctx.VotingEndAlertedProposals[pctx.ChainName] == nil {
		pctx.VotingEndAlertedProposals[pctx.ChainName] = make(map[string]bool)
	}
	pctx.VotingEndAlertedProposals[pctx.ChainName][proposal.ProposalID] = true

	err = h.Services.FirestoreHandler.SaveAlertedProposals(ctx, proposals.CollectionNameVotingEndAlerted, pctx.VotingEndAlertedProposals)
	if err != nil {
		log.Printf("Error saving voting end alerted proposals: %v", err)
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
