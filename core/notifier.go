package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Notifier defines the behavioral contract for sending high-priority alerts to external systems (e.g. Telegram/Slack/Webhooks).
type Notifier interface {
	SendAlert(alert AlertData) error
}

// WebhookNotifier dispatches alert notifications asynchronously to a configured URL endpoint.
type WebhookNotifier struct {
	webhookURL string
	client     *http.Client
	mu         sync.Mutex
}

// NewWebhookNotifier initializes a new WebhookNotifier.
func NewWebhookNotifier(webhookURL string, timeout time.Duration) *WebhookNotifier {
	return &WebhookNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SendAlert sends the alert structured JSON payload to the webhook URL.
func (wn *WebhookNotifier) SendAlert(alert AlertData) error {
	if wn.webhookURL == "" {
		return nil
	}

	wn.mu.Lock()
	defer wn.mu.Unlock()

	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert webhook payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), wn.client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", wn.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Sananti-Security-Notifier/v7.0")

	resp, err := wn.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned non-success code: %d", resp.StatusCode)
	}

	return nil
}
