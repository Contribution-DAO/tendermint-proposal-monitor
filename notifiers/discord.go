package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DiscordNotifier struct {
	WebhookURL string
}

type DiscordMessage struct {
	Content string `json:"content"`
}

func (d *DiscordNotifier) SendAlert(message string) error {
	msg := DiscordMessage{Content: message}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send discord message: %s", resp.Status)
	}

	return nil
}
