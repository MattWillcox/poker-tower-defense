package game

import (
	"math"
	"time"

	"realtime-game-backend/internal/models"
)

// Tower types
const (
	BasicTower  = "basic"
	SplashTower = "splash"
	SniperTower = "sniper"
	SlowTower   = "slow"
)

// Tower costs
var towerCosts = map[string]int{
	BasicTower:  50,
	SplashTower: 100,
	SniperTower: 150,
	SlowTower:   75,
}

// Tower ranges
var towerRanges = map[string]float64{
	BasicTower:  100,
	SplashTower: 75,
	SniperTower: 200,
	SlowTower:   100,
}

// Tower damages
var towerDamages = map[string]int{
	BasicTower:  10,
	SplashTower: 5,  // Splash damage to multiple enemies
	SniperTower: 30, // High damage to single enemy
	SlowTower:   5,  // Low damage but slows enemies
}

// Tower attack speeds (attacks per second)
var towerSpeeds = map[string]float64{
	BasicTower:  1.0,
	SplashTower: 0.5,
	SniperTower: 0.5,
	SlowTower:   1.5,
}

// CreateTower creates a new tower
func CreateTower(playerID, towerType string, x, y float64) models.Tower {
	return models.Tower{
		ID:       GenerateID(),
		PlayerID: playerID,
		Type:     towerType,
		Level:    1,
		X:        x,
		Y:        y,
		Range:    towerRanges[towerType],
		Damage:   towerDamages[towerType],
		Speed:    towerSpeeds[towerType],
		Cost:     towerCosts[towerType],
		LastShot: 0,
	}
}

// UpgradeTower upgrades a tower to the next level
func UpgradeTower(tower models.Tower) models.Tower {
	tower.Level++
	tower.Range *= 1.2
	tower.Damage = int(float64(tower.Damage) * 1.5)
	tower.Speed *= 1.2
	tower.Cost = int(float64(tower.Cost) * 1.5)
	return tower
}

// GetTowerUpgradeCost returns the cost to upgrade a tower
func GetTowerUpgradeCost(tower models.Tower) int {
	return int(float64(towerCosts[tower.Type]) * math.Pow(1.5, float64(tower.Level)))
}

// CanTowerAttack checks if a tower can attack based on its attack speed
func CanTowerAttack(tower models.Tower) bool {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	attackInterval := int64(1000 / tower.Speed) // Convert attacks per second to milliseconds per attack
	return now-tower.LastShot >= attackInterval
}

// UpdateTowerLastShot updates the last shot timestamp for a tower
func UpdateTowerLastShot(tower *models.Tower) {
	tower.LastShot = time.Now().UnixNano() / int64(time.Millisecond)
}

// GetTowerTargets gets the targets for a tower
func GetTowerTargets(tower models.Tower, enemies []models.Enemy) []models.Enemy {
	var targets []models.Enemy

	for _, enemy := range enemies {
		if !enemy.Active {
			continue
		}

		// Calculate distance between tower and enemy
		distance := math.Sqrt(math.Pow(tower.X-enemy.X, 2) + math.Pow(tower.Y-enemy.Y, 2))

		// Check if enemy is in range
		if distance <= tower.Range {
			targets = append(targets, enemy)

			// For non-splash towers, only target the first enemy in range
			if tower.Type != SplashTower {
				break
			}
		}
	}

	return targets
}

// ApplyTowerDamage applies damage from a tower to enemies
func ApplyTowerDamage(tower models.Tower, enemies []models.Enemy) []models.Enemy {
	targets := GetTowerTargets(tower, enemies)
	if len(targets) == 0 {
		return enemies
	}

	// Update the last shot timestamp
	UpdateTowerLastShot(&tower)

	// Apply damage to targets
	for i, target := range targets {
		for j, enemy := range enemies {
			if enemy.ID == target.ID {
				enemies[j].Health -= tower.Damage

				// Apply slow effect for slow towers
				if tower.Type == SlowTower {
					enemies[j].Speed *= 0.7 // Reduce speed by 30%
				}

				// Check if enemy is dead
				if enemies[j].Health <= 0 {
					enemies[j].Active = false
				}

				// For non-splash towers, only damage the first target
				if tower.Type != SplashTower {
					break
				}
			}
		}

		// For non-splash towers, only damage the first target
		if tower.Type != SplashTower && i == 0 {
			break
		}
	}

	return enemies
}

// GenerateID generates a unique ID
func GenerateID() string {
	return time.Now().Format("20060102150405.000000000")
}
