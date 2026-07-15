package services

import (
	"github.com/rafapasa/sales-service/internal/domain/models"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductService struct {
	repo *database.ProductRepository
}

func NewProductService(repo *database.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	return s.repo.Create(product)
}

func (s *ProductService) GetAllProducts() ([]*models.Product, error) {
	return s.repo.GetAll()
}

func (s *ProductService) GetProductByID(id primitive.ObjectID) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) UpdateProduct(id primitive.ObjectID, product *models.Product) error {
	return s.repo.Update(id, product)
}

func (s *ProductService) DeleteProduct(id primitive.ObjectID) error {
	return s.repo.Delete(id)
}
