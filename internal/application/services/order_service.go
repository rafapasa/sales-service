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
	publisher *messaging.SalesPublisher
}

func NewOrderService(repo *database.OrderRepository, publisher *messaging.SalesPublisher) *OrderService {
	return &OrderService{repo: repo,
		publisher: publisher}
}

// EnqueueOrder apenas publica o pedido recebido para processamento assíncrono.
func (s *OrderService) EnqueueOrder(ctx context.Context, order *models.Order) (string, error) {
	// Publicar no RabbitMQ
	envelope, err := s.publisher.PublishOrderCreated(ctx, order)
	if err != nil {
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

func (s *OrderService) UpdateOrder(ctx context.Context, id primitive.ObjectID, order *models.Order) error {
	// Publicar no RabbitMQ
	order.Id = id
	_, err := s.publisher.PublishOrderUpdated(ctx, order)
	if err != nil {
		log.Printf("❌ Erro ao publicar evento: %v", err)
		return err
	}

	log.Printf("✅ Pedido de atualização enfileirado para processamento. OrderID: %s", id)
	return nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, id primitive.ObjectID) error {
	// Publicar no RabbitMQ
	_, err := s.publisher.PublishOrderCanceled(ctx, &models.Order{Id: id})
	if err != nil {
		log.Printf("❌ Erro ao publicar evento: %v", err)
		return err
	}

	log.Printf("✅ Pedido de cancelamento enfileirado para processamento. OrderID: %s", id)
	return nil
}
