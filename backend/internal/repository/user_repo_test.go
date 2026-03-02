package repository

import (
	"context"
	"kashino-backend/internal/models"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestUserRepository(t *testing.T) {
	// Skip if no local mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available")
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Skip("MongoDB not available")
	}

	db := client.Database("kashino_test")
	defer db.Drop(context.Background())
	repo := NewUserRepository(db)

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	balance, err := repo.GetBalance(context.Background(), user.ID)
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	}

	if balance != 1000 {
		t.Errorf("Expected balance 1000, got %f", balance)
	}
}
