package utils

import (
	"fmt"
	"strings"
	"time"
)

func FormatTimeLeft(endTime time.Time) string {
	currentTime := time.Now()
	duration := endTime.Sub(currentTime)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	return fmt.Sprintf("%d days %02d hours", days, hours)
}

func GenerateProposalDetailURL(baseURL, chainName, proposalID string) string {
	return fmt.Sprintf("%s/%s/proposals/%s", baseURL, strings.ToLower(chainName), proposalID)
}
