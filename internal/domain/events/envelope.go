package events

import (
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

type MessageEnvelope struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Version       int                    `json:"version"`
	Timestamp     time.Time              `json:"timestamp"`
	Source        string                 `json:"source"`
	CorrelationID string                 `json:"correlation_id"`
	CausationID   string                 `json:"causation_id,omitempty"`
	Payload       json.RawMessage        `json:"payload"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func NewMessageEnvelope(eventType string, payload interface{}) (*MessageEnvelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &MessageEnvelope{
		ID:            uuid.New().String(),
		Type:          eventType,
		Version:       1,
		Timestamp:     time.Now().UTC(),
		Source:        "sales-service",
		CorrelationID: uuid.New().String(),
		Payload:       payloadBytes,
		Metadata: map[string]interface{}{
			"environment": os.Getenv("ENVIRONMENT"),
		},
	}, nil
}

func (e *MessageEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
