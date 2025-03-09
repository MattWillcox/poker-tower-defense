package models

import (
	"time"
)

// GameSession represents a game session
type GameSession struct {
	ID        string     `json:"id"`
	RoomID    string     `json:"roomId"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
	Status    string     `json:"status"` // "active", "completed", "abandoned"
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// GameState represents the state of a game
type GameState struct {
	SessionID   string                  `json:"sessionId"`
	RoomID      string                  `json:"roomId"`
	Round       int                     `json:"round"`
	Phase       string                  `json:"phase"` // "setup", "cards", "towers", "combat", "end"
	Players     map[string]*PlayerState `json:"players"`
	CurrentWave *EnemyWave              `json:"currentWave,omitempty"`
	StartedAt   int64                   `json:"startedAt"`
	UpdatedAt   int64                   `json:"updatedAt"`
	Status      string                  `json:"status"` // "active", "completed", "abandoned"
}

// Card represents a playing card
type Card struct {
	ID     string `json:"id"`
	Suit   string `json:"suit"`   // "hearts", "diamonds", "clubs", "spades"
	Rank   string `json:"rank"`   // "2", "3", ..., "10", "J", "Q", "K", "A"
	Value  int    `json:"value"`  // 2-14
	Held   bool   `json:"held"`   // Whether the card is being held for the next round
	Active bool   `json:"active"` // Whether the card is active in the current hand
}

// Tower represents a defense tower
type Tower struct {
	ID       string  `json:"id"`
	PlayerID string  `json:"playerId"`
	Type     string  `json:"type"`     // "basic", "splash", "sniper", etc.
	Level    int     `json:"level"`    // 1-3
	X        float64 `json:"x"`        // X position
	Y        float64 `json:"y"`        // Y position
	Range    float64 `json:"range"`    // Attack range
	Damage   int     `json:"damage"`   // Damage per hit
	Speed    float64 `json:"speed"`    // Attack speed (attacks per second)
	Cost     int     `json:"cost"`     // Gold cost
	LastShot int64   `json:"lastShot"` // Timestamp of last shot
}

// HandRank represents a poker hand rank
type HandRank struct {
	Type  string `json:"type"`  // "high_card", "pair", "two_pair", "three_of_a_kind", "straight", "flush", "full_house", "four_of_a_kind", "straight_flush", "royal_flush"
	Value int    `json:"value"` // 1-10
	Name  string `json:"name"`  // Human-readable name
}

// PokerHand represents a poker hand
type PokerHand struct {
	Cards    []Card   `json:"cards"`
	Rank     HandRank `json:"rank"`
	PlayerID string   `json:"playerId"`
}
