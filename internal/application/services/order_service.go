package services

import (
	"github.com/rafapasa/sales-service/models"
	"github.com/rafapasa/sales-service/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	return s.repo.Create(order)
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
