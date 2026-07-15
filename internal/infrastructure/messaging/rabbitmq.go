package messaging

import (
	"context"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    exchange string
    mu       sync.RWMutex
}

func NewRabbitMQPublisher(connString, exchange string) (*RabbitMQPublisher, error) {
    conn, err := amqp.Dial(connString)
    if err != nil {
        return nil, err
    }

    channel, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    // Declarar exchange do tipo topic
    err = channel.ExchangeDeclare(
        exchange,      // nome
        "topic",       // tipo
        true,          // durável
        false,         // auto-delete
        false,         // internal
        false,         // no-wait
        nil,           // argumentos
    )
    if err != nil {
        return nil, err
    }

    publisher := &RabbitMQPublisher{
        conn:     conn,
        channel:  channel,
        exchange: exchange,
    }

    // Monitorar conexão em background
    go publisher.monitorConnection()

    return publisher, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
    p.mu.RLock()
    defer p.mu.RUnlock()

    return p.channel.PublishWithContext(ctx,
        p.exchange,   // exchange
        routingKey,   // routing key
        false,        // mandatory
        false,        // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            Timestamp:    time.Now(),
        },
    )
}

func (p *RabbitMQPublisher) Close() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if err := p.channel.Close(); err != nil {
        return err
    }
    return p.conn.Close()
}

// Monitora e reconecta se perder a conexão
func (p *RabbitMQPublisher) monitorConnection() {
    for {
        if p.conn.IsClosed() {
            log.Println("⚠️ Conexão RabbitMQ perdida. Tentando reconectar...")
            // Implementar reconexão
        }
        time.Sleep(10 * time.Second)
    }
}