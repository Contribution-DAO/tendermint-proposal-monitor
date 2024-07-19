package notifiers

import (
	"bytes"
	"fmt"
	"net/http"
)

var MessageBoxColor = 0x00ffff

type DiscordEmbed struct {
	Color       int    `json:"color"`
	Description string `json:"description"`
}

type DiscordMessage struct {
	Content string         `json:"content"`
	TTS     bool           `json:"tts"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

type DiscordNotifier struct {
	WebhookURL string
}

func (dn *DiscordNotifier) SendPayload(payload []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", dn.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

func (dn *DiscordNotifier) SendAlert(message string) error {
	payload := []byte(`{"content": "` + message + `"}`)
	resp, err := dn.SendPayload(payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send discord message: %s", resp.Status)
	}

	return nil
}
