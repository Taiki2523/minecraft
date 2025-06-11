package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Notifier interface {
	Send(message string) error
}

type DiscordNotifier struct {
	WebhookURL string
}

func (d *DiscordNotifier) Send(message string) error {
	payload := map[string]string{"content": message}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx response from Discord: %s", resp.Status)
	}
	return nil
}
