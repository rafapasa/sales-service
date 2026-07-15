package services

import (
	"context"
	"errors"
	"log"

	"github.com/rafapasa/sales-service/internal/domain/events"
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

func (s *OrderService) CreateOrder(ctx context.Context, req *models.Order) (*models.Order, error) {
	// 1. Validações de negócio
	if req.Total <= 0 {
		return nil, errors.New("total amount must be greater than zero")
	}

	// 2. Salvar no MongoDB
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}

	// Convert models.OrderItem to domainevents.OrderItem
	var eventItems []events.OrderItem
	for _, item := range req.Items {
		eventItems = append(eventItems, events.OrderItem{
			ProductID: item.Product.Id.String(),
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	// 3. PUBLICAR EVENTO (NOSSO FOCO!)
	event := events.NewOrderCreatedEvent(
		req.Id.String(),
		req.Customer.Id.String(),
		req.Total,
		eventItems,
	)

	// Criar envelope
	envelope, err := messaging.NewMessageEnvelope(events.EventOrderCreated, event)
	if err != nil {
		log.Printf("❌ Erro ao criar envelope: %v", err)
		// Não falha a operação, apenas loga
		return req, nil
	}

	// Publicar no RabbitMQ
	body, err := envelope.ToJSON()
	if err != nil {
		log.Printf("❌ Erro ao serializar evento: %v", err)
		return req, nil
	}

	if err := s.publisher.Publish(ctx, events.EventOrderCreated, body); err != nil {
		log.Printf("❌ Erro ao publicar evento: %v", err)
		// Não falha a operação, mas DEVERÍAMOS salvar no OUTBOX
		return req, nil
	}

	log.Printf("✅ Evento publicado: %s", events.EventOrderCreated)

	return req, nil
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
