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
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer database.CloseDB()

	// Conectar RabbitMQ (Publisher para publicar o evento `order.created`)
	publisher, err := messaging.NewRabbitMQPublisher(
		cfg.RabbitMQURI,
		"sales.events", // Mesmo exchange da API
	)
	if err != nil {
		log.Fatalf("Erro ao conectar ao RabbitMQ (Publisher): %v", err)
	}
	defer publisher.Close()

	// Inicializar dependências do Worker
	orderRepo := database.NewOrderRepository(db)
	orderProcessor := processors.NewOrderProcessor(orderRepo, publisher)

	// Criar e iniciar o Consumidor
	consumer, err := messaging.NewRabbitMQConsumer(
		cfg.RabbitMQURI,
		"sales.events",         // Exchange
		"orders.process",       // Nome da Fila
		"order.received.v1",    // Routing Key que queremos consumir
		orderProcessor.Process, // A função que vai processar cada mensagem
	)
	if err != nil {
		log.Fatalf("Erro ao criar consumidor RabbitMQ: %v", err)
	}

	// Iniciar o consumo em uma goroutine
	go func() {
		log.Println("🚀 Worker iniciado. Aguardando pedidos...")
		if err := consumer.Start(); err != nil {
			log.Fatalf("Erro fatal no consumidor: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("📴 Desligando worker...")

	// Shutdown com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := consumer.Shutdown(ctx); err != nil {
		log.Printf("Erro ao desligar consumidor: %v", err)
	} else {
		log.Println("✅ Consumidor desligado com sucesso")
	}

	log.Println("✅ Worker desligado com sucesso")
}
