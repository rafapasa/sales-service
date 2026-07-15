# sales-service
Microsserviço de exemplo para venda com MongoDB

# para teste

docker-compose up -d

go run .\cmd\api\main.go

# chamada postman
POST http://localhost:8080/api/v1/orders

{
    "customer": {
        "customer_id": "507f1f77bcf86cd799439011",
        "Nome": "Rafael Pasa",
        "email": "rafapasagmail.com",
        "andress": "Trv Ana ALbrechet 68, centro - Maravilha SC"
    },
    "items": [
        {
            "product": {
                "product_id": "507f1f77bcf86cd799439012",
                "name": "Prod Teste 1",
                "description": "",
                "price": 29.90
            },
            "quantity": 1,
            "price": 25.00,
            "total": 25.00
        }
    ],
    "total": 25.00
}