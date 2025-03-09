package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"realtime-game-backend/internal/game"
	"realtime-game-backend/internal/models"
)

// Client represents a connected websocket client
type Client struct {
	ID          string
	Connection  *websocket.Conn
	Send        chan []byte
	Hub         *Hub
	PlayerID    string
	RoomID      string
	CurrentHand []models.Card
	CurrentDeck []models.Card
	DrawCount   int
	WaveLevel   int // Track the current wave level
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	Clients map[string]*Client

	// Rooms maps room IDs to a set of clients
	Rooms map[string]map[string]*Client

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Inbound messages from the clients
	Broadcast chan *Message

	// Mutex for concurrent access to maps
	Mutex sync.RWMutex
}

// Message represents a message sent between clients
type Message struct {
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"`
	RoomID   string          `json:"roomId,omitempty"`
	SenderID string          `json:"senderId,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

// NewHub creates a new hub instance
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Rooms:      make(map[string]map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message),
		Mutex:      sync.RWMutex{},
	}
}

// Run starts the hub and handles client registration, unregistration, and message broadcasting
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client.ID] = client
			if client.RoomID != "" {
				if _, ok := h.Rooms[client.RoomID]; !ok {
					h.Rooms[client.RoomID] = make(map[string]*Client)
				}
				h.Rooms[client.RoomID][client.ID] = client
			}
			h.Mutex.Unlock()
			log.Printf("Client registered: %s", client.ID)
		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
				if client.RoomID != "" && h.Rooms[client.RoomID] != nil {
					delete(h.Rooms[client.RoomID], client.ID)
					if len(h.Rooms[client.RoomID]) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}
			h.Mutex.Unlock()
			log.Printf("Client unregistered: %s", client.ID)
		case message := <-h.Broadcast:
			h.Mutex.RLock()
			// If the message has a room ID, send it only to clients in that room
			if message.RoomID != "" {
				if clients, ok := h.Rooms[message.RoomID]; ok {
					for _, client := range clients {
						select {
						case client.Send <- encodeMessage(message):
						default:
							close(client.Send)
							h.Mutex.RUnlock()
							h.Mutex.Lock()
							delete(h.Clients, client.ID)
							if h.Rooms[client.RoomID] != nil {
								delete(h.Rooms[client.RoomID], client.ID)
								if len(h.Rooms[client.RoomID]) == 0 {
									delete(h.Rooms, client.RoomID)
								}
							}
							h.Mutex.Unlock()
							h.Mutex.RLock()
						}
					}
				}
			} else {
				// Broadcast to all clients
				for _, client := range h.Clients {
					select {
					case client.Send <- encodeMessage(message):
					default:
						close(client.Send)
						h.Mutex.RUnlock()
						h.Mutex.Lock()
						delete(h.Clients, client.ID)
						if client.RoomID != "" && h.Rooms[client.RoomID] != nil {
							delete(h.Rooms[client.RoomID], client.ID)
							if len(h.Rooms[client.RoomID]) == 0 {
								delete(h.Rooms, client.RoomID)
							}
						}
						h.Mutex.Unlock()
						h.Mutex.RLock()
					}
				}
			}
			h.Mutex.RUnlock()
		}
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles the connection
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		return
	}

	// Extract player ID and room ID from query parameters
	playerID := r.URL.Query().Get("playerId")
	roomID := r.URL.Query().Get("roomId")

	client := &Client{
		ID:         conn.RemoteAddr().String(),
		Connection: conn,
		Send:       make(chan []byte, 256),
		Hub:        h,
		PlayerID:   playerID,
		RoomID:     roomID,
	}

	h.Register <- client

	// Start goroutines for reading and writing messages
	go client.readPump()
	go client.writePump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(512 * 1024) // 512KB max message size
	log.Printf("Starting readPump for client %s (Player: %s, Room: %s)", c.ID, c.PlayerID, c.RoomID)

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		log.Printf("Received raw message from client %s: %s", c.ID, string(message))

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		log.Printf("Unmarshaled message: Type=%s, SenderID=%s, RoomID=%s", msg.Type, msg.SenderID, msg.RoomID)

		// Set sender ID and room ID if not provided in the message
		if msg.SenderID == "" {
			msg.SenderID = c.PlayerID
		}
		if msg.RoomID == "" {
			msg.RoomID = c.RoomID
		}

		// Handle different message types
		switch msg.Type {
		case "deal_cards":
			// Handle deal_cards message
			log.Printf("Handling deal_cards message from %s", msg.SenderID)

			// Check if this is the first, second, or third draw
			if c.DrawCount == 0 {
				// First draw - generate a new deck and deal 5 cards
				log.Printf("First draw for player %s", msg.SenderID)

				// Generate a new deck
				deck := game.NewDeck()
				log.Printf("Generated new deck with %d cards", len(deck))

				// Shuffle the deck
				shuffledDeck := game.ShuffleDeck(deck)

				// Deal 5 cards
				hand, remainingDeck := game.DealCards(shuffledDeck, 5)
				log.Printf("Dealt 5 cards to player %s: %+v", msg.SenderID, hand)

				// Store the hand and deck for future draws
				c.CurrentHand = hand
				c.CurrentDeck = remainingDeck
				c.DrawCount++

				// Evaluate the hand
				handRank := game.EvaluateHand(hand)
				log.Printf("Hand evaluated as: %s (value: %d)", handRank.Name, handRank.Value)

				// Create response payload
				payload := map[string]interface{}{
					"cards":     hand,
					"handRank":  handRank,
					"drawCount": c.DrawCount,
					"maxDraws":  3, // Indicate that 3 draws are allowed
				}

				// Marshal payload to JSON
				payloadJSON, err := json.Marshal(payload)
				if err != nil {
					log.Printf("Error marshaling payload: %v", err)
					continue
				}

				// Create response message
				response := &Message{
					Type:     "cards_dealt",
					Payload:  payloadJSON,
					RoomID:   msg.RoomID,
					SenderID: "server",
				}

				log.Printf("Sending cards_dealt response to room %s", msg.RoomID)

				// Send response back to the client
				c.Hub.Broadcast <- response
			} else if c.DrawCount == 1 || c.DrawCount == 2 {
				// Second or third draw - keep held cards and replace others
				var drawText string
				if c.DrawCount == 1 {
					drawText = "Second"
				} else {
					drawText = "Third"
				}
				log.Printf("%s draw for player %s", drawText, msg.SenderID)

				// Get the current hand and find which cards are held
				var heldCards []models.Card
				var discardCount int

				for _, card := range c.CurrentHand {
					if card.Held {
						heldCards = append(heldCards, card)
						log.Printf("Keeping held card: %s of %s", card.Rank, card.Suit)
					} else {
						discardCount++
						log.Printf("Discarding card: %s of %s", card.Rank, card.Suit)
					}
				}

				// Draw new cards to replace discarded ones
				log.Printf("Drawing %d new cards", discardCount)
				newHand, remainingDeck := game.DealCards(c.CurrentDeck, discardCount)

				// Combine held cards with new cards
				finalHand := append(heldCards, newHand...)
				log.Printf("Final hand: %+v", finalHand)

				// Update the client's hand and deck
				c.CurrentHand = finalHand
				c.CurrentDeck = remainingDeck
				c.DrawCount++

				// Evaluate the final hand
				handRank := game.EvaluateHand(finalHand)
				log.Printf("Hand evaluated as: %s (value: %d)", handRank.Name, handRank.Value)

				// Calculate gold earned if this is the final draw
				var goldEarned int
				if c.DrawCount >= 3 {
					goldEarned = calculateGoldForHand(handRank.Value)
					log.Printf("Player earned %d gold for %s", goldEarned, handRank.Name)
				}

				// Create response payload
				payload := map[string]interface{}{
					"cards":     finalHand,
					"handRank":  handRank,
					"drawCount": c.DrawCount,
					"maxDraws":  3, // Indicate that 3 draws are allowed
				}

				// Add gold earned if this is the final draw
				if c.DrawCount >= 3 {
					payload["goldEarned"] = goldEarned
				}

				// Marshal payload to JSON
				payloadJSON, err := json.Marshal(payload)
				if err != nil {
					log.Printf("Error marshaling payload: %v", err)
					continue
				}

				// Create response message
				response := &Message{
					Type:     "cards_dealt",
					Payload:  payloadJSON,
					RoomID:   msg.RoomID,
					SenderID: "server",
				}

				log.Printf("Sending cards_dealt response to room %s", msg.RoomID)

				// Send response back to the client
				c.Hub.Broadcast <- response
			} else {
				// Reset for a new round
				log.Printf("Resetting for a new round for player %s", msg.SenderID)
				c.DrawCount = 0
				c.CurrentHand = nil
				c.CurrentDeck = nil

				// Handle as first draw
				// Generate a new deck
				deck := game.NewDeck()
				log.Printf("Generated new deck with %d cards", len(deck))

				// Shuffle the deck
				shuffledDeck := game.ShuffleDeck(deck)

				// Deal 5 cards
				hand, remainingDeck := game.DealCards(shuffledDeck, 5)
				log.Printf("Dealt 5 cards to player %s: %+v", msg.SenderID, hand)

				// Store the hand and deck for future draws
				c.CurrentHand = hand
				c.CurrentDeck = remainingDeck
				c.DrawCount++

				// Evaluate the hand
				handRank := game.EvaluateHand(hand)
				log.Printf("Hand evaluated as: %s (value: %d)", handRank.Name, handRank.Value)

				// Create response payload
				payload := map[string]interface{}{
					"cards":     hand,
					"handRank":  handRank,
					"drawCount": c.DrawCount,
					"maxDraws":  3, // Indicate that 3 draws are allowed
				}

				// Marshal payload to JSON
				payloadJSON, err := json.Marshal(payload)
				if err != nil {
					log.Printf("Error marshaling payload: %v", err)
					continue
				}

				// Create response message
				response := &Message{
					Type:     "cards_dealt",
					Payload:  payloadJSON,
					RoomID:   msg.RoomID,
					SenderID: "server",
				}

				log.Printf("Sending cards_dealt response to room %s", msg.RoomID)

				// Send response back to the client
				c.Hub.Broadcast <- response
			}

		case "hold_hand":
			// Handle hold hand message - skip to final draw
			if room, ok := c.Hub.Rooms[c.RoomID]; ok {
				// Check if the player exists in the room
				client, playerExists := room[c.PlayerID]
				if !playerExists {
					log.Printf("Player %s not found in room %s for hold_hand message", c.PlayerID, c.RoomID)
					continue
				}

				// Only process if we're in the card phase and not already at max draws
				if client.DrawCount < 3 {
					// Set draw count to one less than max to trigger final draw
					client.DrawCount = 2

					// Create a deal_cards message to trigger the final draw
					dealMessage := &Message{
						Type:     "deal_cards",
						Payload:  []byte("{}"),
						RoomID:   c.RoomID,
						SenderID: c.PlayerID,
					}

					// Process the deal_cards message
					c.Hub.Broadcast <- dealMessage

					log.Printf("Player %s is holding their hand and skipping to final draw", c.PlayerID)
				}
			} else {
				log.Printf("Room %s not found for hold_hand message from player %s", c.RoomID, c.PlayerID)
			}

		case "hold_card":
			// Handle hold_card message
			var payload struct {
				CardID string `json:"cardId"`
			}

			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error unmarshaling hold_card payload: %v", err)
				continue
			}

			log.Printf("Player %s is holding card %s", msg.SenderID, payload.CardID)

			// Update the held status of the card in the player's hand
			for i, card := range c.CurrentHand {
				if card.ID == payload.CardID {
					c.CurrentHand[i].Held = true
					log.Printf("Marked card %s as held", payload.CardID)
					break
				}
			}

			// Forward the message to all clients in the room
			c.Hub.Broadcast <- &msg

		case "discard_card":
			// Handle discard_card message
			var payload struct {
				CardID string `json:"cardId"`
			}

			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error unmarshaling discard_card payload: %v", err)
				continue
			}

			log.Printf("Player %s is discarding card %s", msg.SenderID, payload.CardID)

			// Update the held status of the card in the player's hand
			for i, card := range c.CurrentHand {
				if card.ID == payload.CardID {
					c.CurrentHand[i].Held = false
					log.Printf("Marked card %s as not held", payload.CardID)
					break
				}
			}

			// Forward the message to all clients in the room
			c.Hub.Broadcast <- &msg

		case "start_wave":
			// Handle start_wave message
			log.Printf("Handling start_wave message from %s", msg.SenderID)

			// Increment wave level
			c.WaveLevel++
			log.Printf("Starting wave level %d for player %s", c.WaveLevel, c.PlayerID)

			// Create a square path for enemies
			path := []models.Point{
				{X: 50, Y: 50},   // Top-left
				{X: 550, Y: 50},  // Top-right
				{X: 550, Y: 450}, // Bottom-right
				{X: 50, Y: 450},  // Bottom-left
				{X: 50, Y: 50},   // Back to top-left (complete the square)
			}

			// Create enemy wave with the square path
			wave := models.EnemyWave{
				ID:      generateID(),
				Round:   c.WaveLevel,
				Level:   c.WaveLevel, // Include level in the wave data
				Path:    path,
				Status:  "active",
				StartAt: time.Now().UnixNano() / int64(time.Millisecond),
			}

			// Generate enemies based on the wave level
			baseEnemyCount := 5 + c.WaveLevel*2 // More enemies in higher waves

			// Add a boss enemy every 5 levels
			hasBoss := c.WaveLevel > 0 && c.WaveLevel%5 == 0

			// Calculate difficulty multipliers based on wave level
			healthMultiplier := 1.0 + float64(c.WaveLevel-1)*0.2 // +20% health per level
			speedMultiplier := 1.0 + float64(c.WaveLevel-1)*0.05 // +5% speed per level
			goldMultiplier := 1.0 + float64(c.WaveLevel-1)*0.1   // +10% gold per level

			for i := 0; i < baseEnemyCount; i++ {
				enemyType := "basic"

				// Add more variety in enemy types as levels progress
				if c.WaveLevel >= 3 && i%4 == 0 {
					enemyType = "fast"
				} else if c.WaveLevel >= 2 && i%6 == 0 {
					enemyType = "tank"
				} else if i%5 == 0 {
					enemyType = "fast"
				} else if i%7 == 0 {
					enemyType = "tank"
				}

				// Base stats for enemy types
				var baseHealth, baseSpeed, baseGold float64

				switch enemyType {
				case "fast":
					baseHealth = 20
					baseSpeed = 1.5
					baseGold = 7
				case "tank":
					baseHealth = 60
					baseSpeed = 0.7
					baseGold = 10
				default: // basic
					baseHealth = 30
					baseSpeed = 1.0
					baseGold = 5
				}

				// Apply difficulty multipliers
				health := int(baseHealth * healthMultiplier)
				speed := baseSpeed * speedMultiplier
				gold := int(baseGold * goldMultiplier)

				// Create enemy at the start of the path
				enemy := models.Enemy{
					ID:        generateID(),
					Type:      enemyType,
					Health:    health,
					MaxHealth: health,
					Speed:     speed,
					Damage:    1,
					Gold:      gold,
					X:         path[0].X,
					Y:         path[0].Y,
					PathIndex: 0,
					Active:    true,
				}

				wave.Enemies = append(wave.Enemies, enemy)
			}

			// Add a boss enemy if this is a boss wave
			if hasBoss {
				bossHealth := int(100 * healthMultiplier)
				bossSpeed := 0.6 * speedMultiplier
				bossGold := int(25 * goldMultiplier)

				boss := models.Enemy{
					ID:        generateID(),
					Type:      "boss",
					Health:    bossHealth,
					MaxHealth: bossHealth,
					Speed:     bossSpeed,
					Damage:    3, // Boss does more damage
					Gold:      bossGold,
					X:         path[0].X,
					Y:         path[0].Y,
					PathIndex: 0,
					Active:    true,
				}

				wave.Enemies = append(wave.Enemies, boss)
				log.Printf("Added boss enemy to wave %d", c.WaveLevel)
			}

			// Create response payload
			payload := map[string]interface{}{
				"wave": wave,
			}

			// Marshal payload to JSON
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error marshaling payload: %v", err)
				continue
			}

			// Create response message
			response := &Message{
				Type:     "wave_started",
				Payload:  payloadJSON,
				RoomID:   msg.RoomID,
				SenderID: "server",
			}

			log.Printf("Sending wave_started response to room %s with %d enemies", msg.RoomID, len(wave.Enemies))

			// Send response back to the client
			c.Hub.Broadcast <- response

			// Reset draw count to allow dealing cards again after the wave
			c.DrawCount = 0

		case "place_tower":
			// Handle place_tower message
			var payload struct {
				TowerType string  `json:"towerType"`
				X         float64 `json:"x"`
				Y         float64 `json:"y"`
			}

			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error unmarshaling place_tower payload: %v", err)
				continue
			}

			log.Printf("Player %s is placing a %s tower at (%.1f, %.1f)", msg.SenderID, payload.TowerType, payload.X, payload.Y)

			// Create a new tower
			tower := models.Tower{
				ID:       generateID(),
				PlayerID: msg.SenderID,
				Type:     payload.TowerType,
				Level:    1,
				X:        payload.X,
				Y:        payload.Y,
				LastShot: 0,
			}

			// Set tower stats based on type
			switch payload.TowerType {
			case "basic":
				tower.Range = 100
				tower.Damage = 10
				tower.Speed = 1.0
				tower.Cost = 50
			case "splash":
				tower.Range = 75
				tower.Damage = 5
				tower.Speed = 0.5
				tower.Cost = 100
			case "sniper":
				tower.Range = 200
				tower.Damage = 30
				tower.Speed = 0.5
				tower.Cost = 150
			case "slow":
				tower.Range = 100
				tower.Damage = 5
				tower.Speed = 1.5
				tower.Cost = 75
			default:
				// Default to basic tower if type is unknown
				tower.Range = 100
				tower.Damage = 10
				tower.Speed = 1.0
				tower.Cost = 50
			}

			// Create response payload
			towerPayload := map[string]interface{}{
				"tower": tower,
			}

			// Marshal payload to JSON
			towerJSON, err := json.Marshal(towerPayload)
			if err != nil {
				log.Printf("Error marshaling tower payload: %v", err)
				continue
			}

			// Create response message
			response := &Message{
				Type:     "tower_placed",
				Payload:  towerJSON,
				RoomID:   msg.RoomID,
				SenderID: "server",
			}

			log.Printf("Sending tower_placed response to room %s", msg.RoomID)

			// Send response back to the client
			c.Hub.Broadcast <- response

		case "upgrade_tower":
			// Handle upgrade_tower message
			var payload struct {
				TowerID string `json:"towerId"`
			}

			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error unmarshaling upgrade_tower payload: %v", err)
				continue
			}

			log.Printf("Player %s is upgrading tower %s", msg.SenderID, payload.TowerID)

			// Create an upgraded tower
			// In a real implementation, you would find the existing tower and upgrade it
			// For this example, we'll create a new upgraded tower

			tower := models.Tower{
				ID:       payload.TowerID,
				PlayerID: msg.SenderID,
				Type:     "basic", // Default type
				Level:    2,       // Upgraded level
				X:        300,     // Default position
				Y:        300,
				Range:    120, // Increased range
				Damage:   15,  // Increased damage
				Speed:    1.2, // Increased attack speed
				Cost:     75,  // Increased cost
				LastShot: 0,
			}

			// Create response payload
			towerPayload := map[string]interface{}{
				"tower": tower,
			}

			// Marshal payload to JSON
			towerJSON, err := json.Marshal(towerPayload)
			if err != nil {
				log.Printf("Error marshaling tower payload: %v", err)
				continue
			}

			// Create response message
			response := &Message{
				Type:     "tower_upgraded",
				Payload:  towerJSON,
				RoomID:   msg.RoomID,
				SenderID: "server",
			}

			log.Printf("Sending tower_upgraded response to room %s", msg.RoomID)

			// Send response back to the client
			c.Hub.Broadcast <- response

		default:
			// Forward other message types to all clients
			c.Hub.Broadcast <- &msg
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer func() {
		c.Connection.Close()
	}()

	log.Printf("Starting writePump for client %s (Player: %s, Room: %s)", c.ID, c.PlayerID, c.RoomID)

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel
				log.Printf("Send channel closed for client %s", c.ID)
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("Sending message to client %s: %s", c.ID, string(message))

			if err := c.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		}
	}
}

// encodeMessage encodes a message to JSON
func encodeMessage(msg *Message) []byte {
	log.Printf("Encoding message: Type=%s, SenderID=%s, RoomID=%s", msg.Type, msg.SenderID, msg.RoomID)
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return []byte{}
	}
	log.Printf("Encoded message: %s", string(data))
	return data
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(clientID, roomID string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	client, ok := h.Clients[clientID]
	if !ok {
		return
	}

	// Remove from current room if any
	if client.RoomID != "" && client.RoomID != roomID {
		if room, ok := h.Rooms[client.RoomID]; ok {
			delete(room, clientID)
			if len(room) == 0 {
				delete(h.Rooms, client.RoomID)
			}
		}
	}

	// Add to new room
	client.RoomID = roomID
	if _, ok := h.Rooms[roomID]; !ok {
		h.Rooms[roomID] = make(map[string]*Client)
	}
	h.Rooms[roomID][clientID] = client
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(clientID, roomID string) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	client, ok := h.Clients[clientID]
	if !ok {
		return
	}

	if client.RoomID == roomID {
		client.RoomID = ""
		if room, ok := h.Rooms[roomID]; ok {
			delete(room, clientID)
			if len(room) == 0 {
				delete(h.Rooms, roomID)
			}
		}
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, message *Message) {
	message.RoomID = roomID
	h.Broadcast <- message
}

// calculateGoldForHand calculates the amount of gold earned based on hand rank
func calculateGoldForHand(handRankValue int) int {
	// Base gold values for each hand rank
	goldValues := map[int]int{
		1:  10,  // High Card
		2:  20,  // Pair
		3:  30,  // Two Pair
		4:  50,  // Three of a Kind
		5:  80,  // Straight
		6:  100, // Flush
		7:  150, // Full House
		8:  200, // Four of a Kind
		9:  300, // Straight Flush
		10: 500, // Royal Flush
	}

	// Get the gold value for the hand rank, default to 10 if not found
	gold, ok := goldValues[handRankValue]
	if !ok {
		gold = 10
	}

	return gold
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
