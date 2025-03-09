package game

import (
	"math"
	"math/rand"
	"time"

	"realtime-game-backend/internal/models"
)

// CreateEnemyWave creates a new enemy wave for a round
func CreateEnemyWave(round int) models.EnemyWave {
	wave := models.EnemyWave{
		ID:      GenerateID(),
		Round:   round,
		Enemies: generateEnemies(round),
		Path:    generatePath(),
		Status:  "pending",
		StartAt: time.Now().Add(5*time.Second).UnixNano() / int64(time.Millisecond),
	}

	return wave
}

// generateEnemies generates enemies for a wave based on the round
func generateEnemies(round int) []models.Enemy {
	var enemies []models.Enemy

	// Base number of enemies - increased scaling
	baseEnemies := 5 + round*3 // Increased from round*2

	// Enemy type probabilities based on round - stronger enemies appear earlier
	basicProb := 1.0
	fastProb := 0.0
	tankProb := 0.0
	bossProb := 0.0

	if round >= 2 { // Reduced from round 3
		basicProb = 0.7
		fastProb = 0.3
	}

	if round >= 4 { // Reduced from round 5
		basicProb = 0.6
		fastProb = 0.3
		tankProb = 0.1
	}

	if round >= 7 { // Reduced from round 10
		basicProb = 0.5
		fastProb = 0.3
		tankProb = 0.15
		bossProb = 0.05
	}

	if round >= 10 { // Added new tier for later rounds
		basicProb = 0.4
		fastProb = 0.3
		tankProb = 0.2
		bossProb = 0.1
	}

	// Generate enemies
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	enemyTypes := models.GetEnemyTypes()

	for i := 0; i < baseEnemies; i++ {
		var enemyType string
		roll := r.Float64()

		switch {
		case roll < bossProb:
			enemyType = "boss"
		case roll < bossProb+tankProb:
			enemyType = "tank"
		case roll < bossProb+tankProb+fastProb:
			enemyType = "fast"
		case roll < bossProb+tankProb+fastProb+basicProb:
			enemyType = "basic"
		default:
			enemyType = "basic" // Fallback
		}

		// Scale enemy health based on round - increased scaling
		healthMultiplier := 1.0 + float64(round-1)*0.2 // Increased from 0.1

		enemy := models.Enemy{
			ID:        GenerateID(),
			Type:      enemyType,
			Health:    int(float64(enemyTypes[enemyType].Health) * healthMultiplier),
			MaxHealth: int(float64(enemyTypes[enemyType].Health) * healthMultiplier),
			Speed:     enemyTypes[enemyType].Speed,
			Damage:    enemyTypes[enemyType].Damage,
			Gold:      enemyTypes[enemyType].Gold,
			X:         0,
			Y:         0,
			PathIndex: 0,
			Active:    true,
		}

		enemies = append(enemies, enemy)
	}

	return enemies
}

// generatePath generates a path for enemies to follow
func generatePath() []models.Point {
	// This is a simplified path generation
	// In a real game, this would be more complex and possibly map-specific
	return []models.Point{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
		{X: 100, Y: 100},
		{X: 200, Y: 100},
		{X: 200, Y: 200},
		{X: 300, Y: 200},
		{X: 300, Y: 300},
		{X: 400, Y: 300},
		{X: 400, Y: 400},
		{X: 500, Y: 400},
	}
}

// UpdateEnemyPositions updates the positions of enemies along the path
func UpdateEnemyPositions(wave models.EnemyWave, deltaTime float64) models.EnemyWave {
	for i, enemy := range wave.Enemies {
		if !enemy.Active {
			continue
		}

		// Get current and next points on the path
		if enemy.PathIndex >= len(wave.Path)-1 {
			// Enemy reached the end of the path
			wave.Enemies[i].Active = false
			continue
		}

		currentPoint := wave.Path[enemy.PathIndex]
		nextPoint := wave.Path[enemy.PathIndex+1]

		// Calculate direction vector
		dx := nextPoint.X - currentPoint.X
		dy := nextPoint.Y - currentPoint.Y

		// Normalize direction vector
		length := distance(currentPoint, nextPoint)
		if length > 0 {
			dx /= length
			dy /= length
		}

		// Calculate movement distance
		moveDistance := enemy.Speed * deltaTime

		// Calculate new position
		newX := enemy.X + dx*moveDistance
		newY := enemy.Y + dy*moveDistance

		// Check if enemy reached or passed the next point
		if distance(models.Point{X: newX, Y: newY}, nextPoint) <= moveDistance {
			// Move to the next point on the path
			wave.Enemies[i].PathIndex++
			wave.Enemies[i].X = nextPoint.X
			wave.Enemies[i].Y = nextPoint.Y
		} else {
			// Update position
			wave.Enemies[i].X = newX
			wave.Enemies[i].Y = newY
		}
	}

	return wave
}

// distance calculates the distance between two points
func distance(p1, p2 models.Point) float64 {
	return math.Sqrt(math.Pow(p2.X-p1.X, 2) + math.Pow(p2.Y-p1.Y, 2))
}

// IsWaveComplete checks if a wave is complete (all enemies are inactive)
func IsWaveComplete(wave models.EnemyWave) bool {
	for _, enemy := range wave.Enemies {
		if enemy.Active {
			return false
		}
	}
	return true
}

// GetActiveEnemies gets all active enemies in a wave
func GetActiveEnemies(wave models.EnemyWave) []models.Enemy {
	var activeEnemies []models.Enemy
	for _, enemy := range wave.Enemies {
		if enemy.Active {
			activeEnemies = append(activeEnemies, enemy)
		}
	}
	return activeEnemies
}

// CalculateWaveDamage calculates the damage done by enemies that reached the end of the path
func CalculateWaveDamage(wave models.EnemyWave) int {
	damage := 0
	for _, enemy := range wave.Enemies {
		if !enemy.Active && enemy.PathIndex >= len(wave.Path)-1 {
			damage += enemy.Damage
		}
	}
	return damage
}

// CalculateWaveGold calculates the gold earned from killing enemies in a wave
func CalculateWaveGold(wave models.EnemyWave) int {
	gold := 0
	for _, enemy := range wave.Enemies {
		if !enemy.Active && enemy.Health <= 0 {
			gold += enemy.Gold
		}
	}
	return gold
}
