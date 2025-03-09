package db

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

// Error definitions
var (
	ErrMissingConnectionString = errors.New("missing database connection string")
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	conn *pgx.Conn
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(ctx context.Context) (*PostgresDB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, ErrMissingConnectionString
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to PostgreSQL")
	return &PostgresDB{conn: conn}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close(ctx context.Context) error {
	return db.conn.Close(ctx)
}

// InitSchema initializes the database schema
func (db *PostgresDB) InitSchema(ctx context.Context) error {
	// Create players table
	_, err := db.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS players (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	// Create game_sessions table
	_, err = db.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS game_sessions (
			id VARCHAR(36) PRIMARY KEY,
			room_id VARCHAR(36) NOT NULL,
			started_at TIMESTAMP NOT NULL DEFAULT NOW(),
			ended_at TIMESTAMP,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	// Create player_sessions table (join table between players and game_sessions)
	_, err = db.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS player_sessions (
			id VARCHAR(36) PRIMARY KEY,
			player_id VARCHAR(36) NOT NULL REFERENCES players(id),
			session_id VARCHAR(36) NOT NULL REFERENCES game_sessions(id),
			score INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			UNIQUE(player_id, session_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create player_stats table
	_, err = db.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS player_stats (
			player_id VARCHAR(36) PRIMARY KEY REFERENCES players(id),
			games_played INTEGER NOT NULL DEFAULT 0,
			games_won INTEGER NOT NULL DEFAULT 0,
			total_score INTEGER NOT NULL DEFAULT 0,
			highest_score INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	// Create high_scores table
	_, err = db.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS high_scores (
			id SERIAL PRIMARY KEY,
			player_name VARCHAR(50) NOT NULL,
			score INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	// Create index on high_scores for faster retrieval
	_, err = db.conn.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_high_scores_score ON high_scores (score DESC)
	`)
	if err != nil {
		return err
	}

	log.Println("✅ Database schema initialized")
	return nil
}

// CreatePlayer creates a new player
func (db *PostgresDB) CreatePlayer(ctx context.Context, id, username string) error {
	_, err := db.conn.Exec(ctx, `
		INSERT INTO players (id, username)
		VALUES ($1, $2)
		ON CONFLICT (username) DO NOTHING
	`, id, username)
	if err != nil {
		return err
	}

	// Initialize player stats
	_, err = db.conn.Exec(ctx, `
		INSERT INTO player_stats (player_id)
		VALUES ($1)
		ON CONFLICT (player_id) DO NOTHING
	`, id)

	return err
}

// CreateGameSession creates a new game session
func (db *PostgresDB) CreateGameSession(ctx context.Context, id, roomID string) error {
	_, err := db.conn.Exec(ctx, `
		INSERT INTO game_sessions (id, room_id)
		VALUES ($1, $2)
	`, id, roomID)
	return err
}

// AddPlayerToSession adds a player to a game session
func (db *PostgresDB) AddPlayerToSession(ctx context.Context, id, playerID, sessionID string) error {
	_, err := db.conn.Exec(ctx, `
		INSERT INTO player_sessions (id, player_id, session_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (player_id, session_id) DO NOTHING
	`, id, playerID, sessionID)
	return err
}

// UpdatePlayerScore updates a player's score in a game session
func (db *PostgresDB) UpdatePlayerScore(ctx context.Context, playerID, sessionID string, score int) error {
	_, err := db.conn.Exec(ctx, `
		UPDATE player_sessions
		SET score = $1, updated_at = NOW()
		WHERE player_id = $2 AND session_id = $3
	`, score, playerID, sessionID)
	return err
}

// EndGameSession marks a game session as ended
func (db *PostgresDB) EndGameSession(ctx context.Context, sessionID string) error {
	_, err := db.conn.Exec(ctx, `
		UPDATE game_sessions
		SET ended_at = NOW(), status = 'completed', updated_at = NOW()
		WHERE id = $1
	`, sessionID)
	return err
}

// UpdatePlayerStats updates a player's stats after a game
func (db *PostgresDB) UpdatePlayerStats(ctx context.Context, playerID string, won bool, score int) error {
	_, err := db.conn.Exec(ctx, `
		UPDATE player_stats
		SET 
			games_played = games_played + 1,
			games_won = games_won + CASE WHEN $1 THEN 1 ELSE 0 END,
			highest_score = GREATEST(highest_score, $2),
			total_score = total_score + $2,
			updated_at = NOW()
		WHERE player_id = $3
	`, won, score, playerID)
	return err
}

// GetPlayerStats gets a player's stats
func (db *PostgresDB) GetPlayerStats(ctx context.Context, playerID string) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	row := db.conn.QueryRow(ctx, `
		SELECT 
			games_played, 
			games_won, 
			highest_score, 
			total_score
		FROM player_stats
		WHERE player_id = $1
	`, playerID)

	var gamesPlayed, gamesWon, highestScore, totalScore int
	err := row.Scan(&gamesPlayed, &gamesWon, &highestScore, &totalScore)
	if err != nil {
		return nil, err
	}

	stats["games_played"] = gamesPlayed
	stats["games_won"] = gamesWon
	stats["highest_score"] = highestScore
	stats["total_score"] = totalScore

	if gamesPlayed > 0 {
		stats["win_rate"] = float64(gamesWon) / float64(gamesPlayed)
		stats["average_score"] = float64(totalScore) / float64(gamesPlayed)
	} else {
		stats["win_rate"] = 0.0
		stats["average_score"] = 0.0
	}

	return stats, nil
}

// GetLeaderboard gets the top players by score
func (db *PostgresDB) GetLeaderboard(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	rows, err := db.conn.Query(ctx, `
		SELECT 
			p.id,
			p.username,
			ps.games_played,
			ps.games_won,
			ps.highest_score,
			ps.total_score
		FROM player_stats ps
		JOIN players p ON p.id = ps.player_id
		ORDER BY ps.highest_score DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []map[string]interface{}
	for rows.Next() {
		var id, username string
		var gamesPlayed, gamesWon, highestScore, totalScore int

		err := rows.Scan(&id, &username, &gamesPlayed, &gamesWon, &highestScore, &totalScore)
		if err != nil {
			return nil, err
		}

		entry := map[string]interface{}{
			"id":            id,
			"username":      username,
			"games_played":  gamesPlayed,
			"games_won":     gamesWon,
			"highest_score": highestScore,
			"total_score":   totalScore,
		}

		if gamesPlayed > 0 {
			entry["win_rate"] = float64(gamesWon) / float64(gamesPlayed)
			entry["average_score"] = float64(totalScore) / float64(gamesPlayed)
		} else {
			entry["win_rate"] = 0.0
			entry["average_score"] = 0.0
		}

		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

// GetHighScores retrieves the top high scores from the database
func (db *PostgresDB) GetHighScores(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10 // Default to top 10 if not specified
	}

	rows, err := db.conn.Query(ctx, `
		SELECT player_name, score, created_at::text
		FROM high_scores
		ORDER BY score DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var highScores []map[string]interface{}
	for rows.Next() {
		var playerName string
		var score int
		var createdAt string

		if err := rows.Scan(&playerName, &score, &createdAt); err != nil {
			return nil, err
		}

		highScore := map[string]interface{}{
			"name":       playerName,
			"score":      score,
			"created_at": createdAt,
		}
		highScores = append(highScores, highScore)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return highScores, nil
}

// SaveHighScore saves a high score to the database and returns whether it's a top score
func (db *PostgresDB) SaveHighScore(ctx context.Context, playerName string, score int) (bool, error) {
	// Check if this score is in the top 10
	var lowestTopScore int
	var count int

	err := db.conn.QueryRow(ctx, `
		SELECT COUNT(*) FROM high_scores
	`).Scan(&count)
	if err != nil {
		return false, err
	}

	isHighScore := false

	if count < 10 {
		// Less than 10 scores, so this is automatically a high score
		isHighScore = true
	} else {
		// Check if this score is higher than the lowest top 10 score
		err = db.conn.QueryRow(ctx, `
			SELECT MIN(score) FROM (
				SELECT score FROM high_scores
				ORDER BY score DESC
				LIMIT 10
			) AS top_scores
		`).Scan(&lowestTopScore)
		if err != nil {
			return false, err
		}

		isHighScore = score > lowestTopScore
	}

	if isHighScore {
		// Insert the new high score
		_, err = db.conn.Exec(ctx, `
			INSERT INTO high_scores (player_name, score)
			VALUES ($1, $2)
		`, playerName, score)
		if err != nil {
			return false, err
		}

		// If we have more than 10 high scores, delete the lowest ones
		if count >= 10 {
			_, err = db.conn.Exec(ctx, `
				DELETE FROM high_scores
				WHERE id IN (
					SELECT id FROM high_scores
					ORDER BY score ASC
					LIMIT (SELECT COUNT(*) - 10 FROM high_scores)
				)
			`)
			if err != nil {
				return false, err
			}
		}
	}

	return isHighScore, nil
}
