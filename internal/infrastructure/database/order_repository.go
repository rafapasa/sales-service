package database

import (
	"context"

	"github.com/rafapasa/sales-service/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *OrderRepository) Create(order *models.Order) error {
	_, err := r.collection.InsertOne(context.Background(), order)
	return err
}

func (r *OrderRepository) GetAll() ([]*models.Order, error) {
	var orders []*models.Order
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) GetByID(id primitive.ObjectID) (*models.Order, error) {
	var order models.Order
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Update(id primitive.ObjectID, order *models.Order) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": order})
	return err
}

func (r *OrderRepository) Delete(id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
