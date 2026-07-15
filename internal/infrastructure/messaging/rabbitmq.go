package messaging

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	exchange    string
	mu          sync.RWMutex
	connString  string
	isConnected bool
	notifyClose chan *amqp.Error
}

func NewRabbitMQPublisher(connString, exchange string) (*RabbitMQPublisher, error) {
	publisher := &RabbitMQPublisher{
		exchange:   exchange,
		connString: connString,
	}

	go publisher.handleReconnect()

	if err := publisher.connect(); err != nil {
		return nil, err
	}

	return publisher, nil
}

func (p *RabbitMQPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Println("🔌 Tentando conectar ao RabbitMQ...")
	conn, err := amqp.Dial(p.connString)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		p.exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	p.conn = conn
	p.channel = ch
	p.isConnected = true
	p.notifyClose = make(chan *amqp.Error)
	p.conn.NotifyClose(p.notifyClose)

	log.Println("✅ Conectado ao RabbitMQ com sucesso!")
	return nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.isConnected {
		return errors.New("falha ao publicar: não conectado ao RabbitMQ")
	}

	return p.channel.PublishWithContext(ctx,
		p.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
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

	if !p.isConnected {
		return nil // Já está fechado
	}

	p.isConnected = false
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}

	close(p.notifyClose)
	return nil
}

// Monitora e reconecta se perder a conexão
func (p *RabbitMQPublisher) handleReconnect() {
	for {
		<-p.notifyClose
		p.mu.Lock()
		p.isConnected = false
		p.mu.Unlock()
		log.Println("⚠️ Conexão RabbitMQ perdida. Tentando reconectar...")

		for {
			time.Sleep(5 * time.Second)
			if err := p.connect(); err == nil {
				break
			}
			log.Println("❌ Falha na reconexão, tentando novamente em 5s")
		}
	}
}
