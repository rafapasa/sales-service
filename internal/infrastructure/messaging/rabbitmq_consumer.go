package messaging

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageHandler func(ctx context.Context, body []byte) error

type RabbitMQConsumer struct {
	connString   string
	exchange     string
	queueName    string
	routingKey   string
	handler      MessageHandler
	conn         *amqp.Connection
	channel      *amqp.Channel
	done         chan bool
	notifyClose  chan *amqp.Error
	isConsuming  bool
}

func NewRabbitMQConsumer(connString, exchange, queueName, routingKey string, handler MessageHandler) (*RabbitMQConsumer, error) {
	consumer := &RabbitMQConsumer{
		connString: connString,
		exchange:   exchange,
		queueName:  queueName,
		routingKey: routingKey,
		handler:    handler,
		done:       make(chan bool),
	}
	return consumer, nil
}

func (c *RabbitMQConsumer) connect() error {
	log.Println("🔌 Tentando conectar ao RabbitMQ como consumidor...")
	conn, err := amqp.Dial(c.connString)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	// Declara a fila
	_, err = ch.QueueDeclare(
		c.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Faz o bind da fila com a exchange e a routing key
	err = ch.QueueBind(
		c.queueName,
		c.routingKey,
		c.exchange,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	c.conn = conn
	c.channel = ch
	c.notifyClose = make(chan *amqp.Error)
	c.conn.NotifyClose(c.notifyClose)

	log.Println("✅ Consumidor conectado e fila configurada!")
	return nil
}

func (c *RabbitMQConsumer) Start() error {
	if err := c.connect(); err != nil {
		return err
	}

	go c.handleReconnect()

	deliveries, err := c.channel.Consume(
		c.queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	c.isConsuming = true
	log.Printf("▶️ Iniciando consumo da fila '%s'", c.queueName)

	for {
		select {
		case <-c.done:
			log.Println("⏹️ Consumo interrompido.")
			return nil
		case d := <-deliveries:
			err := c.handler(context.Background(), d.Body)
			if err != nil {
				// Nack - Rejeita a mensagem e a devolve para a fila
				d.Nack(false, true)
			} else {
				// Ack - Confirma o processamento
				d.Ack(false)
			}
		}
	}
}

func (c *RabbitMQConsumer) handleReconnect() {
	<-c.notifyClose
	log.Println("⚠️ Conexão do consumidor perdida. Tentando reconectar e reiniciar o consumo...")
	c.isConsuming = false
	for {
		time.Sleep(5 * time.Second)
		if err := c.Start(); err == nil {
			break
		}
		log.Println("❌ Falha na reconexão do consumidor, tentando novamente em 5s")
	}
}

func (c *RabbitMQConsumer) Shutdown(ctx context.Context) error {
	if !c.isConsuming {
		return nil
	}
	close(c.done)
	c.channel.Close()
	return c.conn.Close()
}