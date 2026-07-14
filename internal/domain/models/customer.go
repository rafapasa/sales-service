package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Customer struct {
	Id      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name    string             `bson:"name" json:"name"`
	Email   string             `bson:"email" json:"email"`
	Address string             `bson:"address" json:"address"`
}
