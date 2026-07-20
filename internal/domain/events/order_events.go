package events

import (
	"time"
)

// OrderCreatedEvent evento de pedido criado
type OrderCreatedEvent struct {
	OrderID      string    `json:"order_id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	TotalAmount  float64   `json:"total_amount"`
	Items        []Item    `json:"items"`
	CreatedAt    time.Time `json:"created_at"`
}

// OrderUpdatedEvent evento de pedido atualizado
type OrderUpdatedEvent struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderCanceledEvent evento de pedido cancelado
type OrderCanceledEvent struct {
	OrderID    string    `json:"order_id"`
	Reason     string    `json:"reason"`
	CanceledAt time.Time `json:"canceled_at"`
}

// PaymentProcessedEvent evento de pagamento processado
type PaymentProcessedEvent struct {
	PaymentID   string    `json:"payment_id"`
	OrderID     string    `json:"order_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

// Item representa um item do pedido
type Item struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

func NewOrderCreatedEvent(orderID, customerID string, total float64, items []Item) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		OrderID:     orderID,
		CustomerID:  customerID,
		TotalAmount: total,
		Items:       items,
	}
}
