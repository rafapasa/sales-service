package messaging

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectionManager gerencia a conexão com RabbitMQ
type ConnectionManager struct {
	conn       *amqp.Connection
	connString string
	mu         sync.RWMutex
	notifyChan chan *amqp.Error
}

var (
	globalManager *ConnectionManager
	managerOnce   sync.Once
)

// GetConnectionManager retorna o singleton do gerenciador
func GetConnectionManager(connString string) *ConnectionManager {
	managerOnce.Do(func() {
		globalManager = &ConnectionManager{
			connString: connString,
		}
	})
	return globalManager
}

// Connect estabelece ou reestabelece a conexão
func (cm *ConnectionManager) Connect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Se já tem conexão válida, retorna
	if cm.conn != nil && !cm.conn.IsClosed() {
		return nil
	}

	log.Println("🔌 Conectando ao RabbitMQ...")

	conn, err := amqp.Dial(cm.connString)
	if err != nil {
		return err
	}

	cm.conn = conn
	cm.notifyChan = conn.NotifyClose(make(chan *amqp.Error))

	// Inicia monitoramento
	go cm.monitorConnection()

	log.Println("✅ Conectado ao RabbitMQ com sucesso")
	return nil
}

// GetConnection retorna a conexão ativa
func (cm *ConnectionManager) GetConnection() (*amqp.Connection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.conn == nil || cm.conn.IsClosed() {
		return nil, &ConnectionError{msg: "conexão não está ativa"}
	}

	return cm.conn, nil
}

// GetChannel cria um canal ativo
func (cm *ConnectionManager) GetChannel() (*amqp.Channel, error) {
	// Primeiro, tenta pegar a conexão
	conn, err := cm.GetConnection()
	if err != nil {
		// Tenta reconectar
		if err := cm.Connect(); err != nil {
			return nil, err
		}
		conn, _ = cm.GetConnection()
	}

	// Tenta abrir canal
	ch, err := conn.Channel()
	if err != nil {
		// Se falhar, tenta reconectar
		if err := cm.Connect(); err != nil {
			return nil, err
		}
		conn, _ = cm.GetConnection()
		return conn.Channel()
	}

	return ch, nil
}

// monitorConnection monitora perda de conexão
func (cm *ConnectionManager) monitorConnection() {
	for {
		err := <-cm.notifyChan
		if err != nil {
			log.Printf("⚠️ Conexão RabbitMQ perdida: %v", err)

			cm.mu.Lock()
			cm.conn = nil
			cm.mu.Unlock()

			// Tenta reconectar automaticamente
			for {
				time.Sleep(5 * time.Second)
				if err := cm.Connect(); err == nil {
					log.Println("✅ Reconectado com sucesso!")
					break
				}
				log.Printf("❌ Falha na reconexão: %v", err)
			}
		}
	}
}

// IsConnected verifica se está conectado
func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.conn != nil && !cm.conn.IsClosed()
}

// ConnectionError erro de conexão
type ConnectionError struct {
	msg string
}

func (e *ConnectionError) Error() string {
	return e.msg
}
