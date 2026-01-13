package queue

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ManagementClient wraps RabbitMQ management API calls.
type ManagementClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
}

// QueueInfo represents minimal fields from management API.
type QueueInfo struct {
	Name            string `json:"name"`
	Messages        int    `json:"messages"`
	MessagesReady   int    `json:"messages_ready"`
	MessagesUnacked int    `json:"messages_unacknowledged"`
}

// ListQueues returns queue infos filtered by prefix (optional).
func (m *ManagementClient) ListQueues(prefix string) ([]QueueInfo, error) {
	if m.Client == nil {
		m.Client = &http.Client{Timeout: 5 * time.Second}
	}
	url := fmt.Sprintf("%s/api/queues", m.BaseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if m.Username != "" {
		req.SetBasicAuth(m.Username, m.Password)
	}
	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("management api %d: %s", resp.StatusCode, string(body))
	}

	var queues []QueueInfo
	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		return nil, err
	}

	if prefix == "" {
		return queues, nil
	}
	var filtered []QueueInfo
	for _, q := range queues {
		if len(q.Name) >= len(prefix) && q.Name[:len(prefix)] == prefix {
			filtered = append(filtered, q)
		}
	}
	return filtered, nil
}
