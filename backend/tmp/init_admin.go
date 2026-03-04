package main_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Use localhost if running outside docker, or 127.0.0.1
	// The docker-compose shows mongodb is on 27018
	mongoURI := "mongodb://localhost:27018"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	db := client.Database("kashino")
	collection := db.Collection("users")

	username := "admin"
	password := "Admin@123!"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	user := bson.M{
		"username": username,
		"password": string(hashedPassword),
		"role":     "admin",
		"email":    "admin@kashino.com",
		"balance":  0,
		"tier":     "VIP",
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"username": username}
	update := bson.M{"$set": user}

	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin user created/updated successfully!")
}
