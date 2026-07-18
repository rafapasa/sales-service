package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rafapasa/rabbitmq-common/client"
	"github.com/rafapasa/rabbitmq-common/queue"
)

// SalesConsumer é o consumidor específico do sales-service
type SalesConsumer struct {
	consumer     *client.Consumer
	queueManager *queue.QueueManager
}

// NewSalesConsumer cria um novo consumidor
func NewSalesConsumer(connString string) (*SalesConsumer, error) {
	queueManager := SetupQueueManager()

	conn, err := ConnectRabbitMQ(connString)
	if err != nil {
		return nil, err
	}

	consumer, err := client.NewConsumer(conn, queueManager)
	if err != nil {
		return nil, err
	}

	return &SalesConsumer{
		consumer:     consumer,
		queueManager: queueManager,
	}, nil
}

// StartConsumingOrders inicia o consumo de eventos de pedidos
func (c *SalesConsumer) StartConsumingOrders(processor OrderProcessorInterface) error {
	// Handler para processar pedidos
	handler := func(ctx context.Context, delivery amqp091.Delivery) error {
		var event OrderCreatedEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return err
		}

		log.Printf("📦 Pedido recebido: %s | Cliente: %s | Valor: %.2f",
			event.OrderID,
			event.CustomerName,
			event.TotalAmount,
		)

		// Processa o pedido
		return processor.ProcessOrder(ctx, event)
	}

	// Inicia consumo da fila de pedidos
	log.Printf("📨 Iniciando consumo da fila: %s", QueueSalesOrders)
	return c.consumer.Consume(
		QueueSalesOrders,
		handler,
		5, // 5 workers
	)
}

// StartConsumingPayments inicia o consumo de eventos de pagamentos
func (c *SalesConsumer) StartConsumingPayments(processor PaymentProcessorInterface) error {
	// Handler para processar pagamentos
	handler := func(ctx context.Context, delivery amqp091.Delivery) error {
		var event PaymentProcessedEvent
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return err
		}

		log.Printf("💳 Pagamento recebido: %s | Order: %s | Status: %s",
			event.PaymentID,
			event.OrderID,
			event.Status,
		)

		// Processa o pagamento
		return processor.ProcessPayment(ctx, event)
	}

	// Inicia consumo da fila de pagamentos
	log.Printf("📨 Iniciando consumo da fila: %s", QueueSalesPayments)
	return c.consumer.Consume(
		QueueSalesPayments,
		handler,
		3, // 3 workers
	)
}

// OrderProcessorInterface define a interface para o processador de pedidos
type OrderProcessorInterface interface {
	ProcessOrder(ctx context.Context, event OrderCreatedEvent) error
}

// PaymentProcessorInterface define a interface para o processador de pagamentos
type PaymentProcessorInterface interface {
	ProcessPayment(ctx context.Context, event PaymentProcessedEvent) error
}
