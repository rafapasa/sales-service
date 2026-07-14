package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type OrderItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Product     Product            `bson:"product" json:"product"`
	Description string             `bson:"description" json:"description"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Price       float64            `bson:"price" json:"price"`
	Total       float64            `bson:"total" json:"total"`
}
