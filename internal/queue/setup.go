// projeto-orders/internal/queue/setup.go
package queue

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/rafapasa/rabbitmq-common/queue"
)

// Constantes específicas do projeto de Orders
const (
	QueueOrders      = "orders"
	QueueOrdersDLQ   = "orders.dlq"
	RoutingKeyOrders = "orders"
	ExchangeOrders   = "exchange_orders"
)

// SetupQueueManager configura o gerenciador de filas específico para Orders
func SetupQueueManager() *queue.QueueManager {
	qm := queue.NewQueueManager()

	// Registra a fila principal de Orders
	qm.RegisterQueue(queue.QueueConfig{
		Name:       QueueOrders,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp091.Table{
			"x-max-priority":            10,
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": QueueOrdersDLQ,
		},
		DLQEnabled:    true,
		DLQName:       QueueOrdersDLQ,
		DLQMaxRetries: 3,
	})

	// Registra a DLQ de Orders
	qm.RegisterQueue(queue.QueueConfig{
		Name:       QueueOrdersDLQ,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
		DLQEnabled: false,
	})

	// Registra o mapeamento de routing key para fila
	qm.RegisterRouting(RoutingKeyOrders, QueueOrders)

	return qm
}
