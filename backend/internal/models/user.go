package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Transaction struct {
	Amount        int64              `bson:"amount" json:"amount"`
	Source        string             `bson:"source" json:"source"` // e.g., "initial", "poker_win", "poker_bet"
	BalanceBefore int64              `bson:"balance_before" json:"balance_before"`
	BalanceAfter  int64              `bson:"balance_after" json:"balance_after"`
	Timestamp     primitive.DateTime `bson:"timestamp" json:"timestamp"`
}

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username         string             `bson:"username" json:"username"`
	Email            string             `bson:"email" json:"email"`
	Password         string             `bson:"password" json:"password"`
	Balance          int64              `bson:"balance" json:"balance"`
	BalanceHistory   []Transaction      `bson:"balance_history" json:"balance_history"`
	Tier             string             `bson:"tier" json:"tier"`
	Exp              int64              `bson:"exp" json:"exp"`
	Level            int                `bson:"level" json:"level"`
	ProfilePicture   string             `bson:"profile_picture,omitempty" json:"profile_picture,omitempty"`
	Role             string             `bson:"role" json:"role"`     // "admin" or "user"
	Status           string             `bson:"status" json:"status"` // "active" or "banned"
	OTP              string             `bson:"otp,omitempty" json:"-"`
	OTPExpiry        primitive.DateTime `bson:"otp_expiry,omitempty" json:"-"`
	ResetToken       string             `bson:"reset_token,omitempty" json:"-"`
	ResetTokenExpiry primitive.DateTime `bson:"reset_token_expiry,omitempty" json:"-"`
	IsVerified       bool               `bson:"is_verified" json:"is_verified"`
	BannedUntil      primitive.DateTime `bson:"banned_until,omitempty" json:"banned_until,omitempty"`
}

// CalculateLevel returns the level based on accumulated EXP.
// Every 200 EXP = 1 level, starting at level 1.
func CalculateLevel(exp int64) int {
	return 1 + int(exp/200)
}

// GetTierFromLevel returns the tier name based on level.
// Wood(1-4), Bronze(5-9), Silver(10-14), Gold(15-19), Fire(20-24), Diamond(25+)
func GetTierFromLevel(level int) string {
	switch {
	case level >= 25:
		return "Diamond"
	case level >= 20:
		return "Fire"
	case level >= 15:
		return "Gold"
	case level >= 10:
		return "Silver"
	case level >= 5:
		return "Bronze"
	default:
		return "Wood"
	}
}

// GetExpForNextLevel returns the total EXP needed to reach the next level.
func GetExpForNextLevel(level int) int64 {
	return int64(level) * 200
}

type UserStats struct {
	TotalWon      int64   `json:"total_won"`
	TotalLost     int64   `json:"total_lost"`
	TotalRefilled int64   `json:"total_refilled"`
	WinRate       float64 `json:"win_rate"`
	TotalGames    int     `json:"total_games"`
}

type Account struct {
	UserID  primitive.ObjectID `bson:"user_id" json:"user_id"`
	Balance int64              `bson:"balance" json:"balance"`
}
