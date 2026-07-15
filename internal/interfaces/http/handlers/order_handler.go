package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rafapasa/sales-service/internal/application/services"
	"github.com/rafapasa/sales-service/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(service *services.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req models.Order
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// A lógica agora é apenas enfileirar o pedido
	correlationID, err := h.service.EnqueueOrder(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Retorna 202 Accepted com um ID para rastreamento
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":        "Order received and is being processed",
		"correlation_id": correlationID,
	})
}

func (h *OrderHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.service.GetAllOrders()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	id, _ := primitive.ObjectIDFromHex(c.Params("id"))
	order, err := h.service.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
	}
	return c.JSON(order)
}
