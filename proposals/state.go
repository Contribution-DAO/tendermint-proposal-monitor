package proposals

import (
	"context"
	"fmt"
	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/utils"

	"cloud.google.com/go/firestore"
	// "google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LastCheckedProposals struct {
	Proposals []utils.ProposalKV `firestore:"proposals"`
}

type AlertedProposals struct {
	Proposals []utils.OuterKV `firestore:"proposals"`
}

const (
	CollectionNameLastChecked      = "last_checked_proposals"
	CollectionNameAlertedProposals = "alerted_proposals"
	CollectionNameVotingEndAlerted = "voting_end_alerted_proposals"
)

type FirestoreHandler struct {
	FirestoreClient *firestore.Client
	credentialsFile string
	projectID       string
	databaseID      string
	collectionName  string
}

func New(cfg *config.Configurations) (*FirestoreHandler, error) {
	ctx := context.Background()
	client, err := firestore.NewClientWithDatabase(ctx, cfg.Storage.ProjectID, cfg.Storage.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %v", err)
	}

	return &FirestoreHandler{
		FirestoreClient: client,
		credentialsFile: cfg.Storage.CredentialsPath,
		projectID:       cfg.Storage.ProjectID,
		databaseID:      cfg.Storage.DatabaseID,
		collectionName:  cfg.Storage.CollectionName,
	}, nil
}

func (c *FirestoreHandler) getFirestoreClient() *firestore.Client {
	return c.FirestoreClient
}

func (c *FirestoreHandler) GetLastCheckedProposalIDs(ctx context.Context) (map[string]int, error) {
	client := c.getFirestoreClient()

	doc := client.Collection(c.collectionName).Doc(CollectionNameLastChecked)
	var lastChecked LastCheckedProposals
	dsnap, err := doc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return make(map[string]int), nil
		}
		return nil, err
	}

	err = dsnap.DataTo(&lastChecked)
	if err != nil {
		return nil, err
	}

	return utils.SliceToMap(lastChecked.Proposals), nil
}

func (c *FirestoreHandler) SaveLastCheckedProposalIDs(ctx context.Context, docID string, lastChecked map[string]int) error {
	client := c.getFirestoreClient()

	doc := client.Collection(c.collectionName).Doc(docID)
	entity := LastCheckedProposals{Proposals: utils.MapToSlice(lastChecked)}
	_, err := doc.Set(ctx, entity)
	return err
}

func (c *FirestoreHandler) GetAlertedProposals(ctx context.Context, docID string) (map[string]map[string]bool, error) {
	client := c.getFirestoreClient()

	doc := client.Collection(c.collectionName).Doc(docID)
	var entity AlertedProposals
	dsnap, err := doc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return make(map[string]map[string]bool), nil
		}
		return nil, err
	}

	err = dsnap.DataTo(&entity)
	if err != nil {
		return nil, err
	}

	return utils.NestedSliceToMap(entity.Proposals), nil
}

func (c *FirestoreHandler) SaveAlertedProposals(ctx context.Context, docID string, alertedProposals map[string]map[string]bool) error {
	client := c.getFirestoreClient()

	doc := client.Collection(c.collectionName).Doc(docID)
	entity := AlertedProposals{Proposals: utils.MapToNestedSlice(alertedProposals)}
	_, err := doc.Set(ctx, entity)
	return err
}

func (c *FirestoreHandler) InitState() (map[string]int, map[string]map[string]bool, map[string]map[string]bool, error) {
	ctx := context.Background()

	lastChecked, err := c.GetLastCheckedProposalIDs(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading last checked proposal IDs, defaulting to empty: %v", err)
	}

	alertedProposals, err := c.GetAlertedProposals(ctx, CollectionNameAlertedProposals)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading alerted proposals, defaulting to empty: %v", err)
	}

	votingEndAlertedProposals, err := c.GetAlertedProposals(ctx, CollectionNameVotingEndAlerted)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error loading voting end alerted proposals, defaulting to empty: %v", err)
	}

	return lastChecked, alertedProposals, votingEndAlertedProposals, nil
}
