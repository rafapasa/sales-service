package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	apiURL         = "http://localhost:8080/api/v1/orders"
	mongoURI       = "mongodb://admin:adminpassword@localhost:27017"
	dbName         = "sales_db"
	collectionName = "orders"
	testTimeout    = 20 * time.Second
)

// Estruturas que espelham o JSON da requisição.
// Em um projeto real, poderiam vir de um pacote compartilhado.
type Customer struct {
	CustomerID string `json:"customer_id"`
	Name       string `json:"Nome"`
	Email      string `json:"email"`
	Address    string `json:"andress"`
}

type Product struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Item struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Total    float64 `json:"total"`
}

type OrderPayload struct {
	Customer Customer `json:"customer"`
	Items    []Item   `json:"items"`
	Total    float64  `json:"total"`
}

type OrderInDB struct {
	ID       primitive.ObjectID `bson:"_id"`
	Customer Customer           `bson:"customer"`
	Total    float64            `bson:"total"`
}

func main() {
	log.Println("▶️  Iniciando teste de caixa preta para o sales-service...")

	// 1. Preparar os dados do pedido
	// Usamos um ID de cliente único para facilitar a busca no banco
	uniqueCustomerID := fmt.Sprintf("test-customer-%d", time.Now().UnixNano())
	orderPayload := OrderPayload{
		Customer: Customer{
			CustomerID: uniqueCustomerID,
			Name:       "Cliente de Teste Blackbox",
			Email:      "teste@exemplo.com",
			Address:    "Rua dos Testes, 123",
		},
		Items: []Item{
			{
				Product: Product{
					ProductID: "prod-abc-123",
					Name:      "Produto de Teste",
					Price:     100.0,
				},
				Quantity: 2,
				Price:    95.0,
				Total:    190.0,
			},
		},
		Total: 190.0,
	}

	// 2. Enviar a requisição HTTP para criar o pedido
	log.Printf("📤 Enviando requisição para criar pedido para o cliente: %s", uniqueCustomerID)
	payloadBytes, err := json.Marshal(orderPayload)
	if err != nil {
		log.Fatalf("❌ Falha ao serializar o payload do pedido: %v", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatalf("❌ Falha ao enviar requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		log.Fatalf("❌ Status code inesperado: Recebido %d, esperado %d", resp.StatusCode, http.StatusAccepted)
	}
	log.Println("✅ Requisição HTTP enviada com sucesso (Status 202 Accepted).")

	// 3. Conectar ao MongoDB para verificação
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("❌ Falha ao conectar ao MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database(dbName).Collection(collectionName)

	// 4. Verificar se o worker processou e salvou o pedido
	log.Println("⏳ Aguardando o worker processar a mensagem e salvar no banco de dados...")
	var foundOrder OrderInDB
	// Tenta verificar algumas vezes, pois o processamento é assíncrono
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second) // Dê tempo para o worker
		err = collection.FindOne(ctx, bson.M{"customer.customer_id": uniqueCustomerID}).Decode(&foundOrder)
		if err == nil {
			break // Encontrou!
		}
	}

	if err != nil {
		log.Fatalf("❌ Pedido não encontrado no banco de dados para o cliente '%s' após várias tentativas. Erro: %v", uniqueCustomerID, err)
	}

	log.Printf("✅ Pedido encontrado no banco de dados! ID: %s", foundOrder.ID.Hex())

	// Validações adicionais
	if foundOrder.Total != orderPayload.Total {
		log.Fatalf("❌ Verificação falhou: Total do pedido no banco (%.2f) é diferente do esperado (%.2f)", foundOrder.Total, orderPayload.Total)
	}
	log.Println("✅ Total do pedido verificado com sucesso.")

	// 5. Limpeza: Remover o pedido de teste do banco de dados
	log.Printf("🧹 Limpando dados de teste (removendo pedido %s)...", foundOrder.ID.Hex())
	_, err = collection.DeleteOne(ctx, bson.M{"_id": foundOrder.ID})
	if err != nil {
		log.Printf("⚠️ Falha ao limpar o pedido de teste do banco de dados: %v", err)
	} else {
		log.Println("✅ Dados de teste removidos.")
	}

	log.Println("🎉 Teste de caixa preta concluído com sucesso!")
}
