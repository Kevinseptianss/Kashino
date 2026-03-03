package main

import (
	"context"
	"fmt"
	"kashino-backend/internal/api/handlers"
	"kashino-backend/internal/repository"
	"kashino-backend/internal/websocket"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDB() (*mongo.Client, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27018"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

func main() {
	client, err := connectDB()
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}
	db := client.Database("kashino")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	pokerRepo := repository.NewPokerRepository(db)

	// Handlers
	userHandler := handlers.NewUserHandler(userRepo)

	// WebSocket Hub
	hub := websocket.NewHub(userRepo, pokerRepo)
	go hub.Run()

	// Routes
	http.HandleFunc("/signup", userHandler.CreateUser)
	http.HandleFunc("/signin", userHandler.SignIn)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Backend server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
