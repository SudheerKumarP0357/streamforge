package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CosmosDB struct {
	client *mongo.Client
	db     *mongo.Database
}

type WatchHistory struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	VideoID         string             `bson:"video_id" json:"video_id"`
	WatchedAt       time.Time          `bson:"watched_at" json:"watched_at"`
	ProgressSeconds int                `bson:"progress_seconds" json:"progress_seconds"`
	Completed       bool               `bson:"completed" json:"completed"`
}

type UserPreferences struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	Theme             string             `bson:"theme" json:"theme"`
	Autoplay          bool               `bson:"autoplay" json:"autoplay"`
	QualityPreference string             `bson:"quality_preference" json:"quality_preference"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

type WatchEvent struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VideoID         string             `bson:"video_id" json:"video_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	EventType       string             `bson:"event_type" json:"event_type"` // play, pause, seek, end
	Timestamp       time.Time          `bson:"timestamp" json:"timestamp"`
	SessionID       string             `bson:"session_id" json:"session_id"`
	PositionSeconds int                `bson:"position_seconds" json:"position_seconds"`
}

func NewCosmos(ctx context.Context, connectionString, dbName string) (*CosmosDB, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("cosmos connection string is empty")
	}

	clientOptions := options.Client().ApplyURI(connectionString)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cosmos db: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping cosmos db: %w", err)
	}

	database := client.Database(dbName)

	return &CosmosDB{
		client: client,
		db:     database,
	}, nil
}

func (c *CosmosDB) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

func (c *CosmosDB) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *CosmosDB) SaveWatchEvent(ctx context.Context, event WatchEvent) error {
	collection := c.db.Collection("video_events")

	// Ensure ObjectID is generated if absent
	if event.ID == primitive.NilObjectID {
		event.ID = primitive.NewObjectID()
	}

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to save watch event: %w", err)
	}

	return nil
}

func (c *CosmosDB) UpsertWatchHistory(ctx context.Context, history WatchHistory) error {
	collection := c.db.Collection("watch_history")

	filter := bson.M{
		"user_id":  history.UserID,
		"video_id": history.VideoID,
	}

	update := bson.M{
		"$set": bson.M{
			"watched_at":       history.WatchedAt,
			"progress_seconds": history.ProgressSeconds,
			"completed":        history.Completed,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert watch history: %w", err)
	}

	return nil
}

func (c *CosmosDB) GetWatchHistory(ctx context.Context, userID string, limit int) ([]WatchHistory, error) {
	collection := c.db.Collection("watch_history")

	filter := bson.M{"user_id": userID}
	
	findOptions := options.Find().
		SetSort(bson.D{{Key: "watched_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find watch history: %w", err)
	}
	defer cursor.Close(ctx)

	var history []WatchHistory
	if err = cursor.All(ctx, &history); err != nil {
		return nil, fmt.Errorf("failed to decode watch history: %w", err)
	}

	return history, nil
}

func (c *CosmosDB) UpsertUserPreferences(ctx context.Context, prefs UserPreferences) error {
	collection := c.db.Collection("user_preferences")

	filter := bson.M{"user_id": prefs.UserID}

	update := bson.M{
		"$set": bson.M{
			"theme":              prefs.Theme,
			"autoplay":           prefs.Autoplay,
			"quality_preference": prefs.QualityPreference,
			"updated_at":         prefs.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert user preferences: %w", err)
	}

	return nil
}

func (c *CosmosDB) GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	collection := c.db.Collection("user_preferences")

	filter := bson.M{"user_id": userID}

	var prefs UserPreferences
	err := collection.FindOne(ctx, filter).Decode(&prefs)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return &prefs, nil
}
