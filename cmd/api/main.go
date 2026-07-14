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
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/rafapasa/sales-service/config"
	"github.com/rafapasa/sales-service/database"
	"github.com/rafapasa/sales-service/handlers"
	"github.com/rafapasa/sales-service/repository"
	"github.com/rafapasa/sales-service/services"
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

	// Inicializar dependências
	customerRepo := repository.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepo)
	customerHandler := handlers.NewCustomerHandler(customerService)

	productRepo := repository.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	orderRepo := repository.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Configurar rotas

	// Criar servidor
	// Criar app Fiber
	app := fiber.New(fiber.Config{
		AppName:      "GopherStore",
		ServerHeader: "Fiber",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		// Error handler customizado
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log do erro
			log.Printf("Erro: %v", err)

			// Retornar erro em JSON
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Middlewares globais
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(fiberrecover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Grupo de rotas da API
	api := app.Group("/api/v1")

	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
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

	// Iniciar servidor em goroutine
	go func() {
		log.Printf("🚀 Servidor Fiber iniciado na porta %s", cfg.ServerPort)
		if err := app.Listen(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("📴 Desligando servidor...")

	// Shutdown com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Erro ao desligar servidor: %v", err)
	}

	log.Println("✅ Servidor desligado com sucesso")
}
