package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/rafapasa/rabbitmq-common/client"
	"github.com/rafapasa/sales-service/internal/application/services"
	"github.com/rafapasa/sales-service/internal/config"
	"github.com/rafapasa/sales-service/internal/infrastructure/database"
	"github.com/rafapasa/sales-service/internal/infrastructure/messaging"
	"github.com/rafapasa/sales-service/internal/interfaces/http/handlers"
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

	// 1. Criar o ConnectionManager (uma vez por aplicação)
	connManager, err := client.NewConnectionManager(cfg.RabbitMQURI)

	// ===== CONFIGURAÇÃO DO RABBITMQ =====
	publisher, err := messaging.NewSalesPublisher(connManager)
	if err != nil {
		log.Fatalf("❌ Erro ao configurar RabbitMQ: %v", err)
	}
	log.Println("✅ RabbitMQ configurado com sucesso!")

	// ===== INICIALIZAR DEPENDÊNCIAS =====
	customerRepo := database.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepo)
	customerHandler := handlers.NewCustomerHandler(customerService)

	productRepo := database.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	orderRepo := database.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, publisher) // ✅ Passa o publisher
	orderHandler := handlers.NewOrderHandler(orderService)

	// ===== CONFIGURAÇÃO DO FIBER =====
	app := fiber.New(fiber.Config{
		AppName:      "GopherStore - Sales Service",
		ServerHeader: "Fiber",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("❌ Erro: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Middlewares
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	// ===== ROTAS =====
	api := app.Group("/api/v1")

	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	// ===== ROTA DE HEALTH CHECK COM RABBITMQ =====
	app.Get("/health", func(c *fiber.Ctx) error {
		health := fiber.Map{
			"status":  "ok",
			"service": "sales-service",
			"time":    time.Now().Format(time.RFC3339),
		}
		checker := messaging.NewHealthChecker(connManager)
		// Verifica RabbitMQ
		if checker != nil {
			if status, err := checker.CheckHealth(); err != nil {
				health["rabbitmq"] = fiber.Map{
					"status": status,
					"error":  err.Error(),
				}
				// Retorna status 503 se RabbitMQ estiver com problema
				return c.Status(fiber.StatusServiceUnavailable).JSON(health)
			} else {
				health["rabbitmq"] = fiber.Map{
					"status": status,
				}
			}
		}

		return c.JSON(health)
	})

	// Rotas de Cliente
	api.Post("/customers", customerHandler.CreateCustomer)
	api.Get("/customers", customerHandler.GetAllCustomers)
	api.Get("/customers/:id", customerHandler.GetCustomerByID)
	api.Put("/customers/:id", customerHandler.UpdateCustomer)
	api.Delete("/customers/:id", customerHandler.DeleteCustomer)

	// Rotas de Produto
	api.Post("/products", productHandler.CreateProduct)
	api.Get("/products", productHandler.GetAllProducts)
	api.Get("/products/:id", productHandler.GetProductByID)
	api.Put("/products/:id", productHandler.UpdateProduct)
	api.Delete("/products/:id", productHandler.DeleteProduct)

	// Rotas de Pedido
	api.Post("/orders", orderHandler.CreateOrder)
	api.Get("/orders", orderHandler.GetAllOrders)
	api.Get("/orders/:id", orderHandler.GetOrderByID)
	// api.Put("/orders/:id/status", orderHandler.UpdateOrderStatus)
	// api.Post("/orders/:id/cancel", orderHandler.CancelOrder)

	// ===== INICIAR SERVIDOR =====
	go func() {
		log.Printf("🚀 Servidor Fiber iniciado na porta %s", cfg.ServerPort)
		if err := app.Listen(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil {
			log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
		}
	}()

	// ===== GRACEFUL SHUTDOWN =====
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("📴 Desligando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("❌ Erro ao desligar servidor: %v", err)
	}

	log.Println("✅ Servidor desligado com sucesso")
}
