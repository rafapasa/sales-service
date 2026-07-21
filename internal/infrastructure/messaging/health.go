package messaging

import (
	"errors"
	"sync"
	"time"

	"github.com/rafapasa/rabbitmq-common/client"
)

// HealthChecker verifica a saúde da conexão RabbitMQ
type HealthChecker struct {
	connManager client.ConnectionManager
	lastCheck   time.Time
	status      string
	mu          sync.RWMutex
}

// NewHealthChecker cria um novo checker
func NewHealthChecker(connManager client.ConnectionManager) *HealthChecker {
	return &HealthChecker{
		connManager: connManager,
		status:      "unknown",
	}
}

// CheckHealth verifica se a conexão está saudável
func (h *HealthChecker) CheckHealth() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastCheck = time.Now()

	if !h.connManager.IsConnected() {
		h.status = "unhealthy"
		return h.status, errors.New("conexão perdida")
	}

	// Tenta abrir um canal para testar
	ch, err := h.connManager.GetChannel()
	if err != nil {
		h.status = "degraded"
		return h.status, err
	}
	defer ch.Close()

	h.status = "healthy"
	return h.status, nil
}

// GetStatus retorna o status atual
func (h *HealthChecker) GetStatus() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}
