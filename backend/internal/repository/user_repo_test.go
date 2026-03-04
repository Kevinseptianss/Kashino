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
		t.Errorf("Expected balance 1000, got %d", balance)
	}

	foundUser, err := repo.GetUser(context.Background(), user.ID)
	if err != nil || foundUser == nil {
		t.Fatalf("Failed to fetch user: %v", err)
	}
	if len(foundUser.BalanceHistory) == 0 {
		t.Fatalf("Expected balance history to be initialized")
	}

	initialTx := foundUser.BalanceHistory[0]
	if initialTx.BalanceBefore != 0 || initialTx.BalanceAfter != 1000 {
		t.Errorf("Initial transaction tracking failed: Before=%d, After=%d", initialTx.BalanceBefore, initialTx.BalanceAfter)
	}

	// Test UpdateBalance
	err = repo.UpdateBalance(context.Background(), user.ID, 500, "test_deposit")
	if err != nil {
		t.Errorf("Failed to update balance: %v", err)
	}

	foundUser, _ = repo.GetUser(context.Background(), user.ID)
	if foundUser.Balance != 1500 {
		t.Errorf("Expected balance 1500, got %d", foundUser.Balance)
	}

	if len(foundUser.BalanceHistory) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(foundUser.BalanceHistory))
	}

	lastTx := foundUser.BalanceHistory[1]
	if lastTx.BalanceBefore != 1000 || lastTx.BalanceAfter != 1500 {
		t.Errorf("UpdateBalance tracking failed: Before=%d, After=%d", lastTx.BalanceBefore, lastTx.BalanceAfter)
	}
}
