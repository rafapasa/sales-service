package events

import (
	"time"

	"github.com/google/uuid"
)

type OrderCreatedEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id"`

	// Dados do evento
	OrderID     string      `json:"order_id"`
	CustomerID  string      `json:"customer_id"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func NewOrderCreatedEvent(orderID, customerID string, total float64, items []OrderItem) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		EventID:       uuid.New().String(),
		EventType:     EventOrderCreated,
		Timestamp:     time.Now().UTC(),
		CorrelationID: uuid.New().String(),
		OrderID:       orderID,
		CustomerID:    customerID,
		TotalAmount:   total,
		Items:         items,
	}
}

func (e *OrderCreatedEvent) GetEventID() string       { return e.EventID }
func (e *OrderCreatedEvent) GetEventType() string     { return e.EventType }
func (e *OrderCreatedEvent) GetTimestamp() time.Time  { return e.Timestamp }
func (e *OrderCreatedEvent) GetCorrelationID() string { return e.CorrelationID }
