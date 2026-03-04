package main

import (
	"context"
	"fmt"
	"kashino-backend/internal/api/handlers"
	"kashino-backend/internal/api/middleware"
	"kashino-backend/internal/mail"
	"kashino-backend/internal/models"
	"kashino-backend/internal/repository"
	"kashino-backend/internal/websocket"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
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
	chatRepo := repository.NewChatRepository(db)

	// WebSocket Hub
	hub := websocket.NewHub(userRepo, pokerRepo, chatRepo)
	go hub.Run()

	// Mail Service
	mailSvc := mail.NewMailService()

	// Handlers
	userHandler := handlers.NewUserHandler(userRepo, mailSvc)
	adminHandler := handlers.NewAdminHandler(userRepo, hub)

	// Ensure admin user exists
	ensureAdminUser(userRepo)

	// Routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	http.HandleFunc("/signup", middleware.CORSMiddleware(userHandler.CreateUser))
	http.HandleFunc("/request-signup-otp", middleware.CORSMiddleware(userHandler.RequestSignupOTP))
	http.HandleFunc("/signin", middleware.CORSMiddleware(userHandler.SignIn))
	http.HandleFunc("/send-otp", middleware.CORSMiddleware(userHandler.SendOTP))
	http.HandleFunc("/forgot-password", middleware.CORSMiddleware(userHandler.ForgotPassword))
	http.HandleFunc("/reset-password", middleware.CORSMiddleware(userHandler.ResetPassword))
	http.HandleFunc("/history/slot", middleware.CORSMiddleware(userHandler.GetSlotHistory))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// Admin Routes
	http.HandleFunc("/admin/stats", middleware.CORSMiddleware(middleware.AdminMiddleware(adminHandler.GetDashboardStats)))
	http.HandleFunc("/admin/users", middleware.CORSMiddleware(middleware.AdminMiddleware(adminHandler.GetUsers)))
	http.HandleFunc("/admin/user", middleware.CORSMiddleware(middleware.AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminHandler.GetUser(w, r)
		} else if r.Method == http.MethodPost {
			adminHandler.UpdateUser(w, r)
		} else if r.Method == http.MethodDelete {
			adminHandler.DeleteUser(w, r)
		}
	})))
	http.HandleFunc("/admin/user/ban", middleware.CORSMiddleware(middleware.AdminMiddleware(adminHandler.BanUser)))
	http.HandleFunc("/admin/user/balance", middleware.CORSMiddleware(middleware.AdminMiddleware(adminHandler.AdjustBalance)))
	http.HandleFunc("/admin/history/poker", middleware.CORSMiddleware(middleware.AdminMiddleware(adminHandler.GetHandHistory)))
	http.HandleFunc("/admin/history/slot", middleware.CORSMiddleware(middleware.AdminMiddleware(adminHandler.GetSlotHistory)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Backend server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func ensureAdminUser(repo *repository.UserRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	admin, _ := repo.FindByUsername(ctx, "admin")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin@123!"), bcrypt.DefaultCost)

	if admin == nil {
		newAdmin := &models.User{
			Username:   "admin",
			Password:   string(hashedPassword),
			Role:       "admin",
			Email:      "admin@kashino.my.id",
			Balance:    0,
			Tier:       "VIP",
			Status:     "active",
			IsVerified: true,
		}
		repo.Create(ctx, newAdmin)
		fmt.Println("Default admin user created.")
	} else {
		// Force reset password and email to ensure credentials work
		admin.Password = string(hashedPassword)
		admin.Email = "admin@kashino.my.id"
		admin.Role = "admin"
		admin.Status = "active"
		admin.IsVerified = true
		repo.Update(ctx, admin)
		fmt.Println("Admin credentials synchronized.")
	}
}
