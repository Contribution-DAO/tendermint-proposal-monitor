package services

import (
	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/proposals"
)

type NewServices struct {
	FirestoreHandler *proposals.FirestoreHandler
	Configurations   *config.Configurations
}

func New(firestore *proposals.FirestoreHandler, configs *config.Configurations) *NewServices {
	return &NewServices{
		FirestoreHandler: firestore,
		Configurations:   configs,
	}
}
