package processors

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rafapasa/sales-service/internal/domain/events"
	"github.com/rafapasa/sales-service/internal/domain/models"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"github.com/rafapasa/sales-service/internal/infrastructure/messaging"
)

type OrderProcessor struct {
	repo      *database.OrderRepository
	publisher *messaging.RabbitMQPublisher
}

func NewOrderProcessor(repo *database.OrderRepository, publisher *messaging.RabbitMQPublisher) *OrderProcessor {
	return &OrderProcessor{
		repo:      repo,
		publisher: publisher,
	}
}

// Process é a função que será chamada pelo consumidor RabbitMQ para cada mensagem.
func (p *OrderProcessor) Process(ctx context.Context, body []byte) error {
	// 1. Decodificar o envelope da mensagem
	var envelope messaging.MessageEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		log.Printf("❌ Erro ao decodificar envelope: %v", err)
		return nil // Retorna nil para não reenfileirar uma mensagem mal formatada (ACK)
	}

	// 2. Decodificar o payload (o pedido em si)
	var order models.Order
	if err := json.Unmarshal(envelope.Payload, &order); err != nil {
		log.Printf("❌ Erro ao decodificar payload do pedido: %v", err)
		return nil // ACK
	}

	log.Printf("Processing order %s", order.Id)

	// 3. Validações de negócio
	if order.Total <= 0 {
		log.Printf("❌ Pedido com total inválido: %s", order.Id)
		return nil // ACK
	}

	// 4. Salvar no MongoDB
	if err := p.repo.Create(&order); err != nil {
		log.Printf("❌ Erro ao salvar pedido no DB: %v", err)
		return err // Retorna erro para reenfileirar (NACK)
	}

	// 5. Publicar evento de sucesso `order.created`
	var eventItems []events.OrderItem
	for _, item := range order.Items {
		eventItems = append(eventItems, events.OrderItem{
			ProductID: item.Product.Id.String(),
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	event := events.NewOrderCreatedEvent(order.Id.String(), order.Customer.Id.String(), order.Total, eventItems)
	successEnvelope, err := messaging.NewMessageEnvelope(events.EventOrderCreated, event)
	if err != nil {
		log.Printf("❌ Erro ao criar envelope de sucesso: %v", err)
		return err // NACK
	}

	successBody, err := successEnvelope.ToJSON()
	if err != nil {
		log.Printf("❌ Erro ao serializar evento de sucesso: %v", err)
		return err // NACK
	}

	if err := p.publisher.Publish(ctx, events.EventOrderCreated, successBody); err != nil {
		log.Printf("❌ Erro ao publicar evento de sucesso: %v", err)
		return err // NACK
	}

	log.Printf("✅ Pedido processado e evento 'order.created' publicado para o pedido: %s", order.Id)
	return nil // Sucesso, envia ACK para a fila
}
