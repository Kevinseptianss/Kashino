package repository

import (
	"context"
	"kashino-backend/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
)

type PokerRepository struct {
	historyCol *mongo.Collection
}

func NewPokerRepository(db *mongo.Database) *PokerRepository {
	return &PokerRepository{
		historyCol: db.Collection("poker_history"),
	}
}

func (r *PokerRepository) LogEvent(ctx context.Context, h models.PokerHistory) error {
	_, err := r.historyCol.InsertOne(ctx, h)
	return err
}
