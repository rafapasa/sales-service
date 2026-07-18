package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rafapasa/sales-service/internal/application/processors"
	"github.com/rafapasa/sales-service/internal/application/services"
	"github.com/rafapasa/sales-service/internal/config"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"github.com/rafapasa/sales-service/internal/infrastructure/messaging"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Configuração
	cfg := config.NewConfig()

	// Conectar ao MongoDB
	db, err := database.ConnectDB(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao MongoDB: %v", err)
	}
	defer database.CloseDB()

	// ===== INICIALIZAR RABBITMQ CONSUMER =====
	consumer, err := messaging.NewSalesConsumer(cfg.RabbitMQURI)
	if err != nil {
		log.Fatalf("❌ Erro ao configurar RabbitMQ Consumer: %v", err)
	}

	// ===== INICIALIZAR SERVIÇOS E PROCESSADORES =====
	orderRepo := database.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, nil) // Publisher não é necessário no worker
	orderProcessor := processors.NewOrderProcessor(orderService)
	paymentProcessor := processors.NewPaymentProcessor(orderService)

	// ===== INICIAR CONSUMO =====
	// Consumir eventos de pedidos
	go func() {
		if err := consumer.StartConsumingOrders(orderProcessor); err != nil {
			log.Fatalf("❌ Erro ao consumir pedidos: %v", err)
		}
	}()

	// Consumir eventos de pagamentos
	go func() {
		if err := consumer.StartConsumingPayments(paymentProcessor); err != nil {
			log.Fatalf("❌ Erro ao consumir pagamentos: %v", err)
		}
	}()

	log.Println("🚀 Worker iniciado. Aguardando mensagens...")

	// ===== GRACEFUL SHUTDOWN =====
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("📴 Desligando worker...")

	// Aguarda processamento finalizar
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aqui você pode adicionar lógica para finalizar processamentos pendentes
	<-ctx.Done()

	log.Println("✅ Worker desligado com sucesso")
}
