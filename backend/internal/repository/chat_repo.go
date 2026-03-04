package repository

import (
	"context"
	"kashino-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatRepository struct {
	chatCol *mongo.Collection
}

func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{
		chatCol: db.Collection("room_chat"),
	}
}

func (r *ChatRepository) SaveMessage(ctx context.Context, msg models.ChatMessage) error {
	_, err := r.chatCol.InsertOne(ctx, msg)
	return err
}

func (r *ChatRepository) GetHistory(ctx context.Context, roomID string, limit int64) ([]models.ChatMessage, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cursor, err := r.chatCol.Find(ctx, bson.M{"room_id": roomID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.ChatMessage
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Reverse to have chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	if messages == nil {
		return []models.ChatMessage{}, nil
	}

	return messages, nil
}
