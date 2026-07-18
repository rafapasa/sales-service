package messaging

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn     *amqp.Connection
	connMu   sync.RWMutex
	connOnce sync.Once
)

// ConnectRabbitMQ estabelece conexão com RabbitMQ com reconexão automática
func ConnectRabbitMQ(connString string) (*amqp.Connection, error) {
	connOnce.Do(func() {
		log.Println("🔌 Conectando ao RabbitMQ...")

		var err error
		conn, err = amqp.Dial(connString)
		if err != nil {
			log.Printf("❌ Falha na conexão inicial: %v", err)
			return
		}

		// Monitora perda de conexão
		go monitorConnection(conn, connString)
	})

	connMu.RLock()
	defer connMu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return nil, ErrNotConnected
	}

	return conn, nil
}

// monitorConnection monitora a conexão e reconecta se necessário
func monitorConnection(conn *amqp.Connection, connString string) {
	notifyClose := conn.NotifyClose(make(chan *amqp.Error))

	for {
		err := <-notifyClose
		if err != nil {
			log.Printf("⚠️ Conexão RabbitMQ perdida: %v. Tentando reconectar...", err)

			connMu.Lock()
			conn = nil
			connMu.Unlock()

			// Tenta reconectar com backoff
			for {
				time.Sleep(5 * time.Second)
				newConn, err := amqp.Dial(connString)
				if err == nil {
					connMu.Lock()
					conn = newConn
					connMu.Unlock()

					// Reinicia o monitoramento
					notifyClose = conn.NotifyClose(make(chan *amqp.Error))
					log.Println("✅ Reconectado ao RabbitMQ com sucesso!")
					break
				}
				log.Printf("❌ Falha na reconexão: %v. Tentando novamente em 5s", err)
			}
		}
	}
}

// ErrNotConnected erro de conexão
var ErrNotConnected = &ConnectionError{msg: "não conectado ao RabbitMQ"}

type ConnectionError struct {
	msg string
}

func (e *ConnectionError) Error() string {
	return e.msg
}
