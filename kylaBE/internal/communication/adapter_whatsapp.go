package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// WhatsAppAdapter sends messages via WhatsApp Cloud API.
type WhatsAppAdapter struct {
	accessToken   string
	phoneNumberID string
	httpClient    *http.Client
}

// NewWhatsAppAdapter constructs a WhatsAppAdapter.
func NewWhatsAppAdapter(accessToken, phoneNumberID string) *WhatsAppAdapter {
	return &WhatsAppAdapter{
		accessToken:   accessToken,
		phoneNumberID: phoneNumberID,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Channel returns "whatsapp".
func (a *WhatsAppAdapter) Channel() string {
	return ChannelWhatsApp
}

// Send dispatches a message via WhatsApp Cloud API.
func (a *WhatsAppAdapter) Send(ctx context.Context, conv *Conversation, msg *Message) error {
	// Parse content JSON to extract text or media.
	var content map[string]interface{}
	if err := json.Unmarshal(msg.Content, &content); err != nil {
		return fmt.Errorf("unmarshal message content: %w", err)
	}

	text, _ := content["text"].(string)
	if text == "" {
		return fmt.Errorf("whatsapp message missing text field")
	}

	// Build WhatsApp API payload.
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                conv.ChannelRef, // recipient phone number
		"type":              "text",
		"text": map[string]string{
			"body": text,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal whatsapp payload: %w", err)
	}

	url := fmt.Sprintf("https://graph.facebook.com/v20.0/%s/messages", a.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		log.Printf("[whatsapp] 4xx error sending msg=%s: %s", msg.ID, string(respBody))
		return fmt.Errorf("whatsapp 4xx: %s", string(respBody))
	}
	if resp.StatusCode >= 500 {
		log.Printf("[whatsapp] 5xx error sending msg=%s: %s", msg.ID, string(respBody))
		return fmt.Errorf("whatsapp 5xx: %s", string(respBody))
	}

	log.Printf("[whatsapp] sent msg=%s to=%s status=%d", msg.ID, conv.ChannelRef, resp.StatusCode)
	return nil
}
