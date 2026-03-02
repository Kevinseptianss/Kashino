package repository

import (
	"context"
	"kashino-backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	if user.Balance == 0 {
		user.Balance = 1000 // Default starting balance
	}
	if user.Tier == "" {
		user.Tier = "VIP Silver"
	}

	// Initialize balance history with the initial deposit
	user.BalanceHistory = []models.Transaction{
		{
			Amount:    user.Balance,
			Source:    "initial",
			Timestamp: primitive.NewDateTimeFromTime(time.Now()),
		},
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateBalance(ctx context.Context, userID primitive.ObjectID, amount float64, source string) error {
	transaction := models.Transaction{
		Amount:    amount,
		Source:    source,
		Timestamp: primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$inc":  bson.M{"balance": amount},
			"$push": bson.M{"balance_history": transaction},
		},
	)
	return err
}

func (r *UserRepository) GetBalance(ctx context.Context, userID primitive.ObjectID) (float64, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (r *UserRepository) GetUser(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
