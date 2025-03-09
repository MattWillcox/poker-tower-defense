package models

// Enemy represents an enemy in the game
type Enemy struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`      // "basic", "fast", "tank", "boss"
	Health    int     `json:"health"`    // Current health
	MaxHealth int     `json:"maxHealth"` // Maximum health
	Speed     float64 `json:"speed"`     // Movement speed
	Damage    int     `json:"damage"`    // Damage to player base
	Gold      int     `json:"gold"`      // Gold reward for killing
	X         float64 `json:"x"`         // X position
	Y         float64 `json:"y"`         // Y position
	PathIndex int     `json:"pathIndex"` // Current index in the path
	Active    bool    `json:"active"`    // Whether the enemy is active
}

// EnemyWave represents a wave of enemies
type EnemyWave struct {
	ID      string  `json:"id"`
	Round   int     `json:"round"`   // Game round
	Level   int     `json:"level"`   // Difficulty level (increases with each wave)
	Enemies []Enemy `json:"enemies"` // Enemies in the wave
	Path    []Point `json:"path"`    // Path for enemies to follow
	Status  string  `json:"status"`  // "pending", "active", "completed"
	StartAt int64   `json:"startAt"` // Timestamp when the wave starts
}

// Point represents a 2D point
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// EnemyType represents an enemy type with base stats
type EnemyType struct {
	Type   string  `json:"type"`
	Health int     `json:"health"`
	Speed  float64 `json:"speed"`
	Damage int     `json:"damage"`
	Gold   int     `json:"gold"`
}

// GetEnemyTypes returns all enemy types
func GetEnemyTypes() map[string]EnemyType {
	return map[string]EnemyType{
		"basic": {
			Type:   "basic",
			Health: 150,
			Speed:  1.0,
			Damage: 2,
			Gold:   6,
		},
		"fast": {
			Type:   "fast",
			Health: 120,
			Speed:  1.7,
			Damage: 2,
			Gold:   9,
		},
		"tank": {
			Type:   "tank",
			Health: 300,
			Speed:  0.8,
			Damage: 3,
			Gold:   12,
		},
		"boss": {
			Type:   "boss",
			Health: 800,
			Speed:  0.6,
			Damage: 5,
			Gold:   25,
		},
	}
}
