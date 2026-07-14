package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description" json:"description"`
	Price         float64            `bson:"price" json:"price"`
	QuantityStock int                `bson:"quantity" json:"quantity"`
}
