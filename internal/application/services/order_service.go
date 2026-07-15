package services

import (
	"context"
	"log"

	"github.com/rafapasa/sales-service/internal/domain/models"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"github.com/rafapasa/sales-service/internal/infrastructure/messaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService struct {
	repo      *database.OrderRepository
	publisher *messaging.RabbitMQPublisher
}

func NewOrderService(repo *database.OrderRepository, publisher *messaging.RabbitMQPublisher) *OrderService {
	return &OrderService{repo: repo,
		publisher: publisher}
}

// EnqueueOrder apenas publica o pedido recebido para processamento assíncrono.
func (s *OrderService) EnqueueOrder(ctx context.Context, order *models.Order) (string, error) {
	const eventType = "order.received.v1"

	// Criar envelope
	envelope, err := messaging.NewMessageEnvelope(eventType, order)
	if err != nil {
		log.Printf("❌ Erro ao criar envelope: %v", err)
		return "", err
	}

	// Publicar no RabbitMQ
	body, err := envelope.ToJSON()
	if err != nil {
		log.Printf("❌ Erro ao serializar evento: %v", err)
		return "", err
	}

	if err := s.publisher.Publish(ctx, eventType, body); err != nil {
		log.Printf("❌ Erro ao publicar evento: %v", err)
		return "", err
	}

	log.Printf("✅ Pedido enfileirado para processamento. CorrelationID: %s", envelope.CorrelationID)
	return envelope.CorrelationID, nil
}

func (s *OrderService) GetAllOrders() ([]*models.Order, error) {
	return s.repo.GetAll()
}

func (s *OrderService) GetOrderByID(id primitive.ObjectID) (*models.Order, error) {
	return s.repo.GetByID(id)
}

func (s *OrderService) UpdateOrder(id primitive.ObjectID, order *models.Order) error {
	return s.repo.Update(id, order)
}

func (s *OrderService) DeleteOrder(id primitive.ObjectID) error {
	return s.repo.Delete(id)
}
