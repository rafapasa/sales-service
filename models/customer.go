package models

type Customer struct {
	Id      string `bson:"_id,omitempty" json:"id,omitempty"`
	Name    string `bson:"name" json:"name"`
	Email   string `bson:"email" json:"email"`
	Address string `bson:"address" json:"address"`
}
