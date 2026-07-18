package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rafapasa/rabbitmq-common/client"
	"github.com/rafapasa/rabbitmq-common/queue"
	"github.com/rafapasa/sales-service/internal/domain/models"
)

// SalesPublisher é o publisher específico do sales-service
type SalesPublisher struct {
	publisher    *client.Publisher
	queueManager *queue.QueueManager
}

// NewSalesPublisher cria um novo publisher para o sales-service
func NewSalesPublisher(connString string) (*SalesPublisher, error) {
	// 1. Configura o gerenciador de filas
	queueManager := SetupQueueManager()

	// 2. Conecta ao RabbitMQ
	conn, err := ConnectRabbitMQ(connString)
	if err != nil {
		return nil, err
	}

	// 3. Cria o publisher genérico
	publisher, err := client.NewPublisher(conn, queueManager)
	if err != nil {
		return nil, err
	}

	return &SalesPublisher{
		publisher:    publisher,
		queueManager: queueManager,
	}, nil
}

// PublishOrderCreated publica um evento de pedido criado
func (p *SalesPublisher) PublishOrderCreated(ctx context.Context, order *models.Order) (*MessageEnvelope, error) {
	const eventType = "order.received.v1"

	// Criar envelope
	envelope, err := NewMessageEnvelope(eventType, order)
	if err != nil {
		log.Printf("❌ Erro ao criar envelope: %v", err)
		return nil, err
	}

	log.Printf("📤 Publicando pedido criado: %s", order.Id) // Use order.ID if that's the correct field
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	// The log below is problematic as envelope.Payload is json.RawMessage, not OrderCreatedEvent directly
	// log.Printf("📤 Publicando pedido criado: %s", &envelope.Payload.(OrderCreatedEvent).OrderID)
	return envelope, p.publisher.Publish(ctx, RoutingKeyOrderCreated, body)
}

// PublishOrderUpdated publica um evento de pedido atualizado
func (p *SalesPublisher) PublishOrderUpdated(ctx context.Context, order OrderUpdatedEvent) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	log.Printf("📤 Publicando pedido atualizado: %s", order.OrderID)
	return p.publisher.Publish(ctx, RoutingKeyOrderUpdated, body)
}

// PublishOrderCanceled publica um evento de pedido cancelado
func (p *SalesPublisher) PublishOrderCanceled(ctx context.Context, order OrderCanceledEvent) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	log.Printf("📤 Publicando pedido cancelado: %s", order.OrderID)
	return p.publisher.Publish(ctx, RoutingKeyOrderCanceled, body)
}

// PublishPaymentProcessed publica um evento de pagamento processado
func (p *SalesPublisher) PublishPaymentProcessed(ctx context.Context, payment PaymentProcessedEvent) error {
	body, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	log.Printf("📤 Publicando pagamento processado: %s", payment.PaymentID)
	return p.publisher.Publish(ctx, RoutingKeyPaymentProcessed, body)
}

// Close fecha a conexão (opcional, a conexão é gerenciada globalmente)
func (p *SalesPublisher) Close() error {
	return nil
}
