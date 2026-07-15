package events

import "time"

const (
	// Order Events
	EventOrderCreated   = "order.created.v1"
	EventOrderUpdated   = "order.updated.v1"
	EventOrderCancelled = "order.cancelled.v1"
	EventOrderPaid      = "order.paid.v1"

	// Customer Events
	EventCustomerCreated = "customer.created.v1"

	// Product Events
	EventProductCreated = "product.created.v1"
	EventProductUpdated = "product.updated.v1"
)

// DomainEvent interface que todos os eventos devem implementar
type DomainEvent interface {
	GetEventID() string
	GetEventType() string
	GetTimestamp() time.Time
	GetCorrelationID() string
}
