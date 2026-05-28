package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"kyla-be/shared/events"
)

// InvokeWebhookActivity makes an outbound HTTP request as a workflow side
// effect. The activity is configured with conservative retry behaviour at the
// Temporal level (set in defaultActivityOptions), but Temporal won't retry on
// non-network errors — the activity returns nil on 4xx so the workflow doesn't
// loop on permanently-bad URLs.
//
// Expected node.Config keys:
//
//	url       string — required
//	method    string — optional, defaults to "POST"
//	headers   map    — optional, header name → value
//	body      any    — optional; arbitrary JSON-serialisable payload
//	include_event bool — optional; if true the trigger event is included as
//	                    "trigger" alongside the body
type InvokeWebhookActivity struct {
	Deps Deps
}

func (a *InvokeWebhookActivity) InvokeWebhook(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	httpClient := a.Deps.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	url, _ := params["url"].(string)
	if url == "" {
		return "", errors.New("invoke_webhook: url is required")
	}
	method, _ := params["method"].(string)
	if method == "" {
		method = http.MethodPost
	}

	payload := map[string]interface{}{}
	if v, ok := params["body"]; ok {
		payload["body"] = v
	}
	if include, _ := params["include_event"].(bool); include {
		payload["trigger"] = event
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("invoke_webhook: marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("invoke_webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if hdrs, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range hdrs {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		// Network errors propagate so Temporal retries.
		return "", fmt.Errorf("invoke_webhook: http: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 500 {
		// Retry-worthy server-side failure.
		return "", fmt.Errorf("invoke_webhook: %s returned %d: %s", url, resp.StatusCode, string(respBody))
	}
	if resp.StatusCode >= 400 {
		// Permanent failure — log and swallow so the workflow doesn't loop.
		return fmt.Sprintf("status=%d body=%s", resp.StatusCode, string(respBody)), nil
	}
	return fmt.Sprintf("status=%d", resp.StatusCode), nil
}
