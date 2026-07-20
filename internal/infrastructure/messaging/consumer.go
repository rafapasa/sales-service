package messaging

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rafapasa/rabbitmq-common/client"
	"github.com/rafapasa/rabbitmq-common/middleware"
	"github.com/rafapasa/sales-service/internal/application/processors"
	"github.com/rafapasa/sales-service/internal/domain/events"
	"github.com/rafapasa/sales-service/internal/domain/models"
)

// SalesConsumer é um wrapper que usa o consumer do rabbitmq-common
type SalesConsumer struct {
	consumer *client.Consumer
}

// NewSalesConsumer cria um novo consumidor
func NewSalesConsumer(connString string) (*SalesConsumer, error) {
	// 1. Gerenciador de conexão
	connManager := GetConnectionManager(connString)
	if err := connManager.Connect(); err != nil {
		return nil, err
	}

	// 2. Gerenciador de filas
	queueManager := SetupQueueManager()

	// 3. Cria o consumer genérico passando o gerenciador de filas e de conexão
	consumer := client.NewConsumer(queueManager, connManager)
	return &SalesConsumer{
		consumer: consumer,
	}, nil
}

// StartConsumingOrders inicia o consumo de pedidos
func (c *SalesConsumer) StartConsumingOrders(processor *processors.OrderProcessor) error {
	// Handler específico para processar mensagens de pedidos
	log.Println("Configurando o handler para a fila de pedidos...")
	handler := func(ctx context.Context, delivery amqp.Delivery) error {
		var envelope events.MessageEnvelope
		if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
			return err
		}

		log.Printf("📦 Recebido: %s | CorrelationID: %s",
			envelope.Type,
			envelope.CorrelationID,
		)

		switch envelope.Type {
		case events.EventOrderCreated, events.EventOrderUpdated: // Pode tratar múltiplos eventos
			var order models.Order
			if err := json.Unmarshal(envelope.Payload, &order); err != nil {
				return err
			}
			return processor.ProcessOrder(ctx, order)
		// Adicionar cases para outros eventos como 'order.updated' e 'order.cancelled'
		default:
			log.Printf("⚠️ Evento desconhecido: %s", envelope.Type)
			return nil
		}
	}

	// Aplica middlewares específicos do projeto (se necessário)
	finalHandler := middleware.Chain(
		handler,
		// middlewares específicos do sales-service
	)

	// Consome com reconexão automática e workers
	return c.consumer.Consume(QueueSalesOrders, finalHandler, 5)
}

// StartConsumingPayments inicia o consumo de pagamentos
// func (c *SalesConsumer) StartConsumingPayments(processor processors.PaymentProcessor) error {
// 	handler := func(ctx context.Context, delivery amqp.Delivery) error {
// 		var envelope Envelope
// 		if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
// 			return err
// 		}

// 		log.Printf("💳 Recebido: %s | CorrelationID: %s",
// 			envelope.EventType,
// 			envelope.CorrelationID,
// 		)

// 		switch envelope.EventType {
// 		case "payment.processed.v1":
// 			var payment PaymentProcessedEvent
// 			payloadBytes, _ := json.Marshal(envelope.Payload)
// 			if err := json.Unmarshal(payloadBytes, &payment); err != nil {
// 				return err
// 			}
// 			return processor.ProcessPayment(ctx, payment)

// 		default:
// 			log.Printf("⚠️ Evento desconhecido: %s", envelope.EventType)
// 			return nil
// 		}
// 	}

// 	return c.consumer.Consume(QueueSalesPayments, handler, 3)
// }
