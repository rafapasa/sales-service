package repository

import (
	"context"

	"github.com/rafapasa/sales-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{
		collection: db.Collection("products"),
	}

}

func (r *ProductRepository) Create(product *models.Product) error {
	_, err := r.collection.InsertOne(context.Background(), product)
	return err
}

func (r *ProductRepository) GetAll() ([]*models.Product, error) {
	var products []*models.Product
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	return products, nil
}

func (r *ProductRepository) GetByID(id string) (*models.Product, error) {
	var product models.Product
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Update(id string, product *models.Product) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": product})
	return err
}

func (r *ProductRepository) Delete(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
