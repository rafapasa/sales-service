package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var db *mongo.Database

func ConnectDB(uri, databaseName string) (*mongo.Database, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(uri)
    
    var err error
    client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }

    // Ping para verificar conexão
    if err = client.Ping(ctx, nil); err != nil {
        return nil, err
    }

    db = client.Database(databaseName)
    log.Println("✅ Conectado ao MongoDB com sucesso!")
    return db, nil
}

func GetDB() *mongo.Database {
    return db
}

func GetCollection(collectionName string) *mongo.Collection {
    return db.Collection(collectionName)
}

func CloseDB() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if client != nil {
        if err := client.Disconnect(ctx); err != nil {
            return err
        }
        log.Println("✅ Desconectado do MongoDB")
    }
    return nil
}