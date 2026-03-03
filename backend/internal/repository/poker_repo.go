package repository

import (
	"context"
	"kashino-backend/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
)

type PokerRepository struct {
	historyCol     *mongo.Collection
	slotHistoryCol *mongo.Collection
}

func NewPokerRepository(db *mongo.Database) *PokerRepository {
	return &PokerRepository{
		historyCol:     db.Collection("poker_history"),
		slotHistoryCol: db.Collection("slot_history"),
	}
}

func (r *PokerRepository) LogEvent(ctx context.Context, h models.PokerHistory) error {
	_, err := r.historyCol.InsertOne(ctx, h)
	return err
}

func (r *PokerRepository) LogSlotEvent(ctx context.Context, s models.SlotHistory) error {
	_, err := r.slotHistoryCol.InsertOne(ctx, s)
	return err
}
