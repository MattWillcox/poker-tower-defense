package db

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

// Error definitions
var (
	ErrMissingRedisURL = errors.New("missing Redis URL")
)

// RedisDB represents a Redis database connection
type RedisDB struct {
	client *redis.Client
}

// NewRedisDB creates a new Redis database connection
func NewRedisDB(ctx context.Context) (*RedisDB, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, ErrMissingRedisURL
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to Redis")
	return &RedisDB{client: client}, nil
}

// Close closes the Redis connection
func (db *RedisDB) Close() error {
	return db.client.Close()
}

// SetGameState sets the game state for a room
func (db *RedisDB) SetGameState(ctx context.Context, roomID string, state interface{}) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return db.client.Set(ctx, "game:"+roomID, data, 24*time.Hour).Err()
}

// GetGameState gets the game state for a room
func (db *RedisDB) GetGameState(ctx context.Context, roomID string, state interface{}) error {
	data, err := db.client.Get(ctx, "game:"+roomID).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, state)
}

// DeleteGameState deletes the game state for a room
func (db *RedisDB) DeleteGameState(ctx context.Context, roomID string) error {
	return db.client.Del(ctx, "game:"+roomID).Err()
}

// SetPlayerCards sets the cards for a player in a room
func (db *RedisDB) SetPlayerCards(ctx context.Context, roomID, playerID string, cards interface{}) error {
	data, err := json.Marshal(cards)
	if err != nil {
		return err
	}

	return db.client.Set(ctx, "cards:"+roomID+":"+playerID, data, 24*time.Hour).Err()
}

// GetPlayerCards gets the cards for a player in a room
func (db *RedisDB) GetPlayerCards(ctx context.Context, roomID, playerID string, cards interface{}) error {
	data, err := db.client.Get(ctx, "cards:"+roomID+":"+playerID).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cards)
}

// SetEnemyWave sets the enemy wave for a room
func (db *RedisDB) SetEnemyWave(ctx context.Context, roomID string, wave interface{}) error {
	data, err := json.Marshal(wave)
	if err != nil {
		return err
	}

	return db.client.Set(ctx, "wave:"+roomID, data, 24*time.Hour).Err()
}

// GetEnemyWave gets the enemy wave for a room
func (db *RedisDB) GetEnemyWave(ctx context.Context, roomID string, wave interface{}) error {
	data, err := db.client.Get(ctx, "wave:"+roomID).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, wave)
}

// SetTowers sets the towers for a player in a room
func (db *RedisDB) SetTowers(ctx context.Context, roomID, playerID string, towers interface{}) error {
	data, err := json.Marshal(towers)
	if err != nil {
		return err
	}

	return db.client.Set(ctx, "towers:"+roomID+":"+playerID, data, 24*time.Hour).Err()
}

// GetTowers gets the towers for a player in a room
func (db *RedisDB) GetTowers(ctx context.Context, roomID, playerID string, towers interface{}) error {
	data, err := db.client.Get(ctx, "towers:"+roomID+":"+playerID).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, towers)
}

// AddPlayerToRoom adds a player to a room
func (db *RedisDB) AddPlayerToRoom(ctx context.Context, roomID, playerID string) error {
	return db.client.SAdd(ctx, "room:"+roomID+":players", playerID).Err()
}

// GetPlayersInRoom gets all players in a room
func (db *RedisDB) GetPlayersInRoom(ctx context.Context, roomID string) ([]string, error) {
	return db.client.SMembers(ctx, "room:"+roomID+":players").Result()
}

// RemovePlayerFromRoom removes a player from a room
func (db *RedisDB) RemovePlayerFromRoom(ctx context.Context, roomID, playerID string) error {
	return db.client.SRem(ctx, "room:"+roomID+":players", playerID).Err()
}

// SetPlayerReady sets a player as ready in a room
func (db *RedisDB) SetPlayerReady(ctx context.Context, roomID, playerID string) error {
	return db.client.SAdd(ctx, "room:"+roomID+":ready", playerID).Err()
}

// GetReadyPlayersInRoom gets all ready players in a room
func (db *RedisDB) GetReadyPlayersInRoom(ctx context.Context, roomID string) ([]string, error) {
	return db.client.SMembers(ctx, "room:"+roomID+":ready").Result()
}

// IsRoomReady checks if all players in a room are ready
func (db *RedisDB) IsRoomReady(ctx context.Context, roomID string) (bool, error) {
	players, err := db.GetPlayersInRoom(ctx, roomID)
	if err != nil {
		return false, err
	}

	readyPlayers, err := db.GetReadyPlayersInRoom(ctx, roomID)
	if err != nil {
		return false, err
	}

	return len(players) > 0 && len(players) == len(readyPlayers), nil
}

// ClearRoomReady clears the ready status for all players in a room
func (db *RedisDB) ClearRoomReady(ctx context.Context, roomID string) error {
	return db.client.Del(ctx, "room:"+roomID+":ready").Err()
}

// PublishGameEvent publishes a game event to a channel
func (db *RedisDB) PublishGameEvent(ctx context.Context, channel string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return db.client.Publish(ctx, channel, data).Err()
}

// SubscribeToGameEvents subscribes to game events on a channel
func (db *RedisDB) SubscribeToGameEvents(ctx context.Context, channel string) *redis.PubSub {
	return db.client.Subscribe(ctx, channel)
}
