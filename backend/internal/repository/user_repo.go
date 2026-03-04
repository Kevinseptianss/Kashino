package repository

import (
	"context"
	"kashino-backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if user.Role == "" {
		user.Role = "user"
	}
	if user.Status == "" {
		user.Status = "active"
	}
	// Note: IsVerified should be set by the handler after OTP check

	// Initialize balance history with the initial deposit
	user.BalanceHistory = []models.Transaction{
		{
			Amount:        user.Balance,
			Source:        "initial",
			BalanceBefore: 0,
			BalanceAfter:  user.Balance,
			Timestamp:     primitive.NewDateTimeFromTime(time.Now()),
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

func (r *UserRepository) UpdateBalance(ctx context.Context, userID primitive.ObjectID, amount int64, source string) error {
	// atomicity matters here
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	balanceBefore := user.Balance
	balanceAfter := balanceBefore + amount

	transaction := models.Transaction{
		Amount:        amount,
		Source:        source,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Timestamp:     primitive.NewDateTimeFromTime(time.Now()),
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$inc":  bson.M{"balance": amount},
			"$push": bson.M{"balance_history": transaction},
		},
	)
	return err
}

func (r *UserRepository) GetBalance(ctx context.Context, userID primitive.ObjectID) (int64, error) {
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

func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *UserRepository) GetSlotHistory(ctx context.Context) ([]models.SlotHistory, error) {
	coll := r.collection.Database().Collection("slot_history")
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(100))
	if err != nil {
		return nil, err
	}
	var history []models.SlotHistory
	err = cursor.All(ctx, &history)
	return history, err
}

func (r *UserRepository) GetPokerHistory(ctx context.Context) ([]models.PokerHistory, error) {
	coll := r.collection.Database().Collection("poker_history")
	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(100))
	if err != nil {
		return nil, err
	}
	var history []models.PokerHistory
	err = cursor.All(ctx, &history)
	return history, err
}

func (r *UserRepository) GetDailyStats(ctx context.Context) (map[string]interface{}, error) {
	// Simple count of users as a start
	totalUsers, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Transaction count (poker + slots)
	slotColl := r.collection.Database().Collection("slot_history")
	totalSlots, _ := slotColl.CountDocuments(ctx, bson.M{})

	pokerColl := r.collection.Database().Collection("poker_history")
	totalPoker, _ := pokerColl.CountDocuments(ctx, bson.M{})

	return map[string]interface{}{
		"total_users":        totalUsers,
		"total_transactions": totalSlots + totalPoker,
	}, nil
}

func (r *UserRepository) UpdateOTP(ctx context.Context, email string, otp string, expiry time.Time) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{
			"$set": bson.M{
				"otp":        otp,
				"otp_expiry": primitive.NewDateTimeFromTime(expiry),
			},
		},
	)
	return err
}

func (r *UserRepository) UpdateResetToken(ctx context.Context, email string, token string, expiry time.Time) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{
			"$set": bson.M{
				"reset_token":        token,
				"reset_token_expiry": primitive.NewDateTimeFromTime(expiry),
			},
		},
	)
	return err
}

func (r *UserRepository) FindByResetToken(ctx context.Context, token string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"reset_token": token}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
