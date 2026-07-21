package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rafapasa/rabbitmq-common/client"
	"github.com/rafapasa/sales-service/internal/application/processors"
	"github.com/rafapasa/sales-service/internal/config"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"github.com/rafapasa/sales-service/internal/infrastructure/messaging"
)

func configLogger() {
	// Configura o formato do log
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Adiciona um prefixo
	log.SetPrefix("[APP] ")

	// Redireciona para um arquivo
	file, _ := os.Create("app.log")
	log.SetOutput(file)
}

func main() {
	// Carregar variáveis de ambiente
	configLogger()
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

	connManager, err := client.NewConnectionManager(cfg.RabbitMQURI)
	if err != nil {
		// O NewConnectionManager agora não retorna erro fatal, apenas loga.
		// A aplicação pode continuar e o manager tentará reconectar.
	}

	// ===== INICIALIZAR RABBITMQ CONSUMER =====
	consumer, err := messaging.NewSalesConsumer(connManager)
	if err != nil {
		log.Fatalf("❌ Erro ao configurar RabbitMQ Consumer: %v", err)
	}

	// ===== INICIALIZAR SERVIÇOS E PROCESSADORES =====
	orderRepo := database.NewOrderRepository(db)
	orderProcessor := processors.NewOrderProcessor(orderRepo)

	// ===== INICIAR CONSUMO =====
	// Consumir eventos de pedidos
	// A função StartConsumingOrders é bloqueante, então a executamos em uma goroutine
	// para não travar o graceful shutdown.
	go func() {
		log.Println("Iniciando consumidor de pedidos...")
		if err := consumer.StartConsumingOrders(orderProcessor); err != nil {
			log.Fatalf("❌ Erro fatal ao iniciar o consumidor de pedidos: %v", err)
		}
	}()

	// Consumir eventos de pagamentos
	// go func() {
	// 	if err := consumer.StartConsumingPayments(paymentProcessor); err != nil {
	// 		log.Fatalf("❌ Erro ao consumir pagamentos: %v", err)
	// 	}
	// }()

	log.Println("🚀 Worker iniciado. Aguardando mensagens...")

	// ===== GRACEFUL SHUTDOWN =====
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("📴 Desligando worker...")

	// Aguarda processamento finalizar
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	consumer.Shutdown()

	// Aqui você pode adicionar lógica para finalizar processamentos pendentes
	<-ctx.Done()

	log.Println("✅ Worker desligado com sucesso")
}
