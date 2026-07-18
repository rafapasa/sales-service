package processors

import (
	"context"
	"log"

	"github.com/rafapasa/sales-service/internal/application/services"
	"github.com/rafapasa/sales-service/internal/infrastructure/messaging"
)

// OrderProcessor processa eventos de pedidos
type OrderProcessor struct {
	orderService *services.OrderService
}

// NewOrderProcessor cria um novo processador de pedidos
func NewOrderProcessor(orderService *services.OrderService) *OrderProcessor {
	return &OrderProcessor{
		orderService: orderService,
	}
}

// ProcessOrder processa um evento de pedido criado
func (p *OrderProcessor) ProcessOrder(ctx context.Context, event messaging.OrderCreatedEvent) error {
	log.Printf("🔄 Processando pedido: %s", event.OrderID)

	// Aqui você pode:
	// 1. Validar dados
	// 2. Atualizar estoque
	// 3. Enviar notificações
	// 4. Chamar outros serviços

	// Exemplo: atualiza status do pedido
	// return p.orderService.UpdateOrderStatus(ctx, event.OrderID, "processing")

	return nil
}

// PaymentProcessor processa eventos de pagamentos
type PaymentProcessor struct {
	orderService *services.OrderService
}

// NewPaymentProcessor cria um novo processador de pagamentos
func NewPaymentProcessor(orderService *services.OrderService) *PaymentProcessor {
	return &PaymentProcessor{
		orderService: orderService,
	}
}

// ProcessPayment processa um evento de pagamento
func (p *PaymentProcessor) ProcessPayment(ctx context.Context, event messaging.PaymentProcessedEvent) error {
	log.Printf("🔄 Processando pagamento: %s", event.PaymentID)

	// Aqui você pode:
	// 1. Atualizar status do pedido para "paid"
	// 2. Enviar nota fiscal
	// 3. Atualizar CRM
	// 4. Liberar entrega

	if event.Status == "approved" {
		// return p.orderService.UpdateOrderStatus(ctx, event.OrderID, "paid")
	}

	return nil
}
