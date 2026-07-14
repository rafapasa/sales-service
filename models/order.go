package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PENDING   = "PENDING"
	APPROVED  = "APPROVED"
	INVOICED  = "INVOICED"
	DELIVERED = "DELIVERED"
)

type Order struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderNo   string             `bson:"order_no" json:"order_no"`
	Customer  Customer           `bson:"customer_id" json:"customer_id"`
	OrderDate time.Time          `bson:"order_date" json:"order_date"`
	Status    string             `bson:"status" json:"status"`
	Total     float64            `bson:"total" json:"total"`
	Items     []OrderItem        `bson:"items" json:"items"`
}
