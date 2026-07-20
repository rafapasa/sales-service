package processors

import (
	"context"
	"log"

	"github.com/rafapasa/sales-service/internal/domain/models"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
)

// OrderProcessor processa eventos de pedidos
type OrderProcessor struct {
	repo *database.OrderRepository
}

// NewOrderProcessor cria um novo processador de pedidos
func NewOrderProcessor(repo *database.OrderRepository) *OrderProcessor {
	return &OrderProcessor{
		repo: repo,
	}
}

// ProcessOrder processa um evento de pedido criado
func (p *OrderProcessor) ProcessOrder(ctx context.Context, order models.Order) error {
	log.Printf("🔄 Processando evento de pedido criado: %s", order.Id.Hex())

	// Lógica de negócio do worker: persistir o pedido no banco de dados.
	if err := p.repo.Create(&order); err != nil {
		log.Printf("❌ Erro ao salvar pedido no banco de dados: %v", err)
		// Retornar o erro fará com que a mensagem seja reenfileirada ou vá para a DLQ,
		// dependendo da configuração do middleware do consumer.
		return err
	}

	log.Printf("✅ Pedido %s salvo no banco de dados com sucesso.", order.Id.Hex())
	return nil
}

// Outros métodos para processar atualizações e cancelamentos podem ser adicionados aqui.
// func (p *OrderProcessor) ProcessOrderUpdate(ctx context.Context, order models.Order) error { ... }
// func (p *OrderProcessor) ProcessOrderCancellation(ctx context.Context, orderID primitive.ObjectID) error { ... }
