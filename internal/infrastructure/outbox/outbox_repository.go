package outbox

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OutboxEvent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	EventID     string             `bson:"event_id"`
	EventType   string             `bson:"event_type"`
	Payload     []byte             `bson:"payload"`
	Status      string             `bson:"status"` // pending, published, failed
	CreatedAt   time.Time          `bson:"created_at"`
	PublishedAt *time.Time         `bson:"published_at,omitempty"`
	RetryCount  int                `bson:"retry_count"`
}

type OutboxRepository struct {
	collection *mongo.Collection
}

func NewOutboxRepository(db *mongo.Database) *OutboxRepository {
	return &OutboxRepository{
		collection: db.Collection("outbox_events"),
	}
}

func (r *OutboxRepository) Save(ctx context.Context, event *OutboxEvent) error {
	event.CreatedAt = time.Now()
	event.Status = "pending"
	_, err := r.collection.InsertOne(ctx, event)
	return err
}
