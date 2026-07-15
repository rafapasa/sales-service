package services

import (
	"github.com/rafapasa/sales-service/internal/domain/models"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerService struct {
	repo *database.CustomerRepository
}

func NewCustomerService(repo *database.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) CreateCustomer(customer *models.Customer) error {
	return s.repo.Create(customer)
}

func (s *CustomerService) GetAllCustomers() ([]*models.Customer, error) {
	return s.repo.GetAll()
}

func (s *CustomerService) GetCustomerByID(id primitive.ObjectID) (*models.Customer, error) {
	return s.repo.GetByID(id)
}

func (s *CustomerService) UpdateCustomer(id primitive.ObjectID, customer *models.Customer) error {
	return s.repo.Update(id, customer)
}

func (s *CustomerService) DeleteCustomer(id primitive.ObjectID) error {
	return s.repo.Delete(id)
}
