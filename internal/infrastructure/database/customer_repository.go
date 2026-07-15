package database

import (
	"context"

	"github.com/rafapasa/sales-service/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CustomerRepository struct {
	collection *mongo.Collection
}

func NewCustomerRepository(db *mongo.Database) *CustomerRepository {
	return &CustomerRepository{
		collection: db.Collection("customers"),
	}
}

func (r *CustomerRepository) Create(customer *models.Customer) error {
	_, err := r.collection.InsertOne(context.Background(), customer)
	return err
}

func (r *CustomerRepository) GetAll() ([]*models.Customer, error) {
	var customers []*models.Customer
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &customers); err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *CustomerRepository) GetByID(id primitive.ObjectID) (*models.Customer, error) {
	var customer models.Customer
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) Update(id primitive.ObjectID, customer *models.Customer) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": customer})
	return err
}

func (r *CustomerRepository) Delete(id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
