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

// SalesConsumer gerencia o consumo de mensagens do RabbitMQ para o sales-service
type SalesConsumer struct {
	consumer    *client.Consumer
	connManager client.ConnectionManager
}

// NewSalesConsumer cria um novo consumidor
func NewSalesConsumer(connectionManager client.ConnectionManager) (*SalesConsumer, error) {
	// O QueueManager é criado aqui porque ele contém as definições de fila específicas deste serviço.
	queueManager := SetupQueueManager()
	return &SalesConsumer{
		connManager: connectionManager,
		consumer:    client.NewConsumer(queueManager, connectionManager),
	}, nil
}

// StartConsumingOrders inicia o consumo de pedidos
func (c *SalesConsumer) StartConsumingOrders(processor *processors.OrderProcessor) error {
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

// func (c *SalesConsumer) startConsuming() error {
// 	deliveries, err := c.channel.Consume(
// 		c.queueName,
// 		"",    // consumer tag
// 		false, // auto-ack
// 		false, // exclusive
// 		false, // no-local
// 		false, // no-wait
// 		nil,   // args
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	c.isConsuming = true
// 	c.notifyCancel = make(chan string)
// 	c.channel.NotifyCancel(c.notifyCancel)

// 	log.Printf("▶️  Iniciando consumo da fila '%s'", c.queueName)

// 	for {
// 		select {
// 		case <-c.done:
// 			log.Println("⏹️  Consumo interrompido.")
// 			return nil
// 		case d := <-deliveries:
// 			c.handleDelivery(d)
// 		}
// 	}
// }

// func (c *SalesConsumer) handleDelivery(delivery amqp.Delivery) {
// 	var envelope events.MessageEnvelope
// 	if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
// 		log.Printf("❌ Erro ao decodificar envelope: %v", err)
// 		delivery.Nack(false, false) // Descarta a mensagem
// 		return
// 	}

// 	log.Printf("📦 Recebido: %s | CorrelationID: %s", envelope.Type, envelope.CorrelationID)

// 	var err error
// 	switch envelope.Type {
// 	case events.EventOrderCreated, events.EventOrderUpdated:
// 		var order models.Order
// 		if err = json.Unmarshal(envelope.Payload, &order); err == nil {
// 			err = c.processor.ProcessOrder(context.Background(), order)
// 		}
// 	default:
// 		log.Printf("⚠️ Evento desconhecido: %s", envelope.Type)
// 	}

// 	if err != nil {
// 		log.Printf("❌ Erro ao processar mensagem: %v. A mensagem será rejeitada.", err)
// 		delivery.Nack(false, false) // Envia para DLQ se configurado
// 	} else {
// 		delivery.Ack(false) // Confirma o processamento
// 	}
// }

// func (c *SalesConsumer) handleReconnect() {
// 	select {
// 	case err := <-c.connManager.notifyChan:
// 		if err != nil {
// 			log.Println("⚠️ Conexão do consumidor perdida. Tentando reconectar...")
// 			c.isConsuming = false
// 			for {
// 				time.Sleep(5 * time.Second)
// 				if err := c.connManager.Connect(); err == nil {
// 					if ch, err := c.connManager.GetChannel(); err == nil {
// 						c.channel = ch
// 						go c.startConsuming()
// 						log.Println("✅ Consumidor reconectado e consumo reiniciado.")
// 						return
// 					}
// 				}
// 				log.Println("❌ Falha na reconexão do consumidor, tentando novamente em 5s")
// 			}
// 		}
// 	case <-c.notifyCancel:
// 		log.Println("Consumo cancelado pelo servidor. Lógica de reconexão pode ser adicionada aqui se necessário.")
// 	}
// }

func (c *SalesConsumer) Shutdown() {
	log.Println("Solicitando parada de todos os consumidores do sales-service...")
	c.consumer.StopAll()
}
