package models

import (
	"time"
)

// Player represents a player in the game
type Player struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PlayerStats represents a player's statistics
type PlayerStats struct {
	PlayerID     string    `json:"playerId"`
	GamesPlayed  int       `json:"gamesPlayed"`
	GamesWon     int       `json:"gamesWon"`
	HighestScore int       `json:"highestScore"`
	TotalScore   int       `json:"totalScore"`
	WinRate      float64   `json:"winRate"`
	AvgScore     float64   `json:"avgScore"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// PlayerSession represents a player's session in a game
type PlayerSession struct {
	ID        string    `json:"id"`
	PlayerID  string    `json:"playerId"`
	SessionID string    `json:"sessionId"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PlayerState represents a player's state in a game
type PlayerState struct {
	PlayerID string  `json:"playerId"`
	Username string  `json:"username"`
	Health   int     `json:"health"`
	Gold     int     `json:"gold"`
	Score    int     `json:"score"`
	Cards    []Card  `json:"cards"`
	Towers   []Tower `json:"towers"`
	IsReady  bool    `json:"isReady"`
	IsActive bool    `json:"isActive"`
	LastSeen int64   `json:"lastSeen"`
}
