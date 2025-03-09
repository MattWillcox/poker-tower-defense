package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"realtime-game-backend/internal/db"
	"realtime-game-backend/internal/ws"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Initialize database connections
	postgresDB, err := db.NewPostgresDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close(ctx)

	// Initialize database schema
	if err := postgresDB.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	redisDB, err := db.NewRedisDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisDB.Close()

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run(ctx)

	// Set up HTTP routes
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", hub.HandleWebSocket)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// High scores API endpoints
	mux.HandleFunc("/api/highscores", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow requests from any origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Origin, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle GET request to retrieve high scores
		if r.Method == "GET" {
			highScores, err := postgresDB.GetHighScores(r.Context(), 10)
			if err != nil {
				log.Printf("Error getting high scores: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Convert to JSON and send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Use encoding/json to marshal the response
			if err := json.NewEncoder(w).Encode(highScores); err != nil {
				log.Printf("Error encoding high scores: %v", err)
			}
			return
		}

		// Handle POST request to save a high score
		if r.Method == "POST" {
			// Parse request body
			var scoreData struct {
				Name  string `json:"name"`
				Score int    `json:"score"`
			}

			if err := json.NewDecoder(r.Body).Decode(&scoreData); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Validate input
			if scoreData.Name == "" || scoreData.Score <= 0 {
				http.Error(w, "Invalid name or score", http.StatusBadRequest)
				return
			}

			// Save high score
			isHighScore, err := postgresDB.SaveHighScore(r.Context(), scoreData.Name, scoreData.Score)
			if err != nil {
				log.Printf("Error saving high score: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := struct {
				IsHighScore bool `json:"isHighScore"`
			}{
				IsHighScore: isHighScore,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			return
		}

		// If we get here, the method is not supported
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		log.Println("🚀 Server running on :3000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	// Create a shutdown context with a timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
