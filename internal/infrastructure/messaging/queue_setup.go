package messaging

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/rafapasa/rabbitmq-common/queue"
)

// Constantes específicas do sales-service
const (
	// Nomes das filas
	QueueSalesOrders      = "sales.orders"
	QueueSalesOrdersDLQ   = "sales.orders.dlq"
	QueueSalesPayments    = "sales.payments"
	QueueSalesPaymentsDLQ = "sales.payments.dlq"

	// Routing Keys
	RoutingKeyOrderCreated     = "sales.order.created"
	RoutingKeyOrderUpdated     = "sales.order.updated"
	RoutingKeyOrderCanceled    = "sales.order.canceled"
	RoutingKeyPaymentProcessed = "sales.payment.processed"
)

// SetupQueueManager configura o gerenciador de filas do sales-service
func SetupQueueManager() *queue.QueueManager {
	qm := queue.NewQueueManager()

	// ===== FILA DE PEDIDOS =====
	qm.RegisterQueue(queue.QueueConfig{
		Name:       QueueSalesOrders,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp091.Table{
			"x-max-priority":            10,
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": QueueSalesOrdersDLQ,
		},
		DLQEnabled:    true,
		DLQName:       QueueSalesOrdersDLQ,
		DLQMaxRetries: 3,
	})

	// DLQ dos pedidos
	qm.RegisterQueue(queue.QueueConfig{
		Name:       QueueSalesOrdersDLQ,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
		DLQEnabled: false,
	})

	// ===== FILA DE PAGAMENTOS =====
	qm.RegisterQueue(queue.QueueConfig{
		Name:       QueueSalesPayments,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp091.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": QueueSalesPaymentsDLQ,
		},
		DLQEnabled:    true,
		DLQName:       QueueSalesPaymentsDLQ,
		DLQMaxRetries: 5,
	})

	// DLQ dos pagamentos
	qm.RegisterQueue(queue.QueueConfig{
		Name:       QueueSalesPaymentsDLQ,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
		DLQEnabled: false,
	})

	// ===== MAPEAMENTO ROUTING KEY -> FILA =====
	qm.RegisterRouting(RoutingKeyOrderCreated, QueueSalesOrders)
	qm.RegisterRouting(RoutingKeyOrderUpdated, QueueSalesOrders)
	qm.RegisterRouting(RoutingKeyOrderCanceled, QueueSalesOrders)
	qm.RegisterRouting(RoutingKeyPaymentProcessed, QueueSalesPayments)

	return qm
}
