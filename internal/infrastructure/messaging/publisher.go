package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rafapasa/rabbitmq-common/client"
	"github.com/rafapasa/rabbitmq-common/queue"
	"github.com/rafapasa/sales-service/internal/domain/events"
	"github.com/rafapasa/sales-service/internal/domain/models"
)

// SalesPublisher é o publisher específico do sales-service
type SalesPublisher struct {
	publisher    *client.Publisher
	queueManager *queue.QueueManager
}

// NewSalesPublisher cria um novo publisher para o sales-service
func NewSalesPublisher(connectionManager client.ConnectionManager) (*SalesPublisher, error) {
	// 1. Configura o gerenciador de filas
	queueManager := SetupQueueManager() // Corretamente busca a configuração do serviço

	// 3. Cria o publisher genérico
	publisher, err := client.NewPublisher(connectionManager, queueManager)
	if err != nil {
		return nil, err
	}

	return &SalesPublisher{
		publisher:    publisher,
		queueManager: queueManager,
	}, nil
}

func (p *SalesPublisher) publish(ctx context.Context, eventType, routingKey string, payload interface{}) (*events.MessageEnvelope, error) {
	envelope, err := events.NewMessageEnvelope(eventType, payload)
	if err != nil {
		log.Printf("❌ Erro ao criar envelope: %v", err)
		return nil, err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		log.Printf("❌ Erro ao serializar envelope: %v", err)
		return nil, err
	}

	log.Printf("📤 Publicando evento '%s' para a routing key '%s'", eventType, routingKey)
	return envelope, p.publisher.Publish(ctx, routingKey, body)
}

// PublishOrderCreated publica um evento de pedido criado
func (p *SalesPublisher) PublishOrderCreated(ctx context.Context, order *models.Order) (*events.MessageEnvelope, error) {
	log.Printf("Publicando criação do pedido: %s", order.Id)
	return p.publish(ctx, events.EventOrderCreated, RoutingKeyOrderCreated, order)
}

// PublishOrderUpdated publica um evento de pedido atualizado
func (p *SalesPublisher) PublishOrderUpdated(ctx context.Context, order *models.Order) (*events.MessageEnvelope, error) {
	log.Printf("Publicando atualização do pedido: %s", order.Id)
	return p.publish(ctx, events.EventOrderUpdated, RoutingKeyOrderUpdated, order)
}

// PublishOrderCanceled publica um evento de pedido cancelado
func (p *SalesPublisher) PublishOrderCanceled(ctx context.Context, order *models.Order) (*events.MessageEnvelope, error) {
	log.Printf("Publicando cancelamento do pedido: %s", order.Id)
	return p.publish(ctx, events.EventOrderCancelled, RoutingKeyOrderCanceled, order)
}

// // PublishPaymentProcessed publica um evento de pagamento processado
// func (p *SalesPublisher) PublishPaymentProcessed(ctx context.Context, payment PaymentProcessedEvent) error {
// 	body, err := json.Marshal(payment)
// 	if err != nil {
// 		return err
// 	}

// 	log.Printf("📤 Publicando pagamento processado: %s", payment.PaymentID)
// 	return p.publisher.Publish(ctx, RoutingKeyPaymentProcessed, body)
// }
