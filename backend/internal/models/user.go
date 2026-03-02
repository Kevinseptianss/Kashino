package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Transaction struct {
	Amount    float64            `bson:"amount" json:"amount"`
	Source    string             `bson:"source" json:"source"` // e.g., "initial", "poker_win", "poker_bet"
	Timestamp primitive.DateTime `bson:"timestamp" json:"timestamp"`
}

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username"`
	Email          string             `bson:"email" json:"email"`
	Password       string             `bson:"password" json:"password"`
	Balance        float64            `bson:"balance" json:"balance"`
	BalanceHistory []Transaction      `bson:"balance_history" json:"balance_history"`
}

type Account struct {
	UserID  primitive.ObjectID `bson:"user_id" json:"user_id"`
	Balance float64            `bson:"balance" json:"balance"`
}
