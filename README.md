# Real-time Game Backend

A real-time game backend server built with Go, WebSockets, Redis, and PostgreSQL. This backend is designed for a tower defense game with poker card mechanics.

## Architecture

```
Frontend (Web & Mobile)
        │
        │ WebSocket (bi-directional real-time communication)
        ▼
┌───────────────────────────────────────┐
│          WebSocket Gateway (Go)       │
│ (Player sessions, rooms, messaging)   │
└──────┬───────────────┬────────────────┘
       │               │
       │               └─── Game State Management (Redis)
       │                       - Active game sessions
       │                       - Temporary state (cards, towers, enemies)
       │
       └─── Game Logic & Persistent Storage
               - Poker Hand Evaluations
               - Game progress data (PostgreSQL)
               - Player profile stats
```

## Features

- Real-time WebSocket communication
- Poker hand evaluation
- Tower defense mechanics
- Enemy wave generation
- Player session management
- Persistent player statistics
- Leaderboards

## Directory Structure

```
realtime-game-backend
├── cmd/
│   └── server/
│       └── main.go              // App entrypoint
├── internal/
│   ├── game/
│   │   ├── cards.go             // Card generation logic
│   │   ├── poker.go             // Poker hand evaluation logic
│   │   ├── waves.go             // Enemy wave spawning logic
│   │   └── towers.go            // Tower management logic
│   │
│   ├── ws/
│   │   └── websocket.go         // WebSocket communication handling
│   │
│   ├── db/
│   │   ├── postgres.go          // Database integration & queries
│   │   └── redis.go             // Redis integration for state caching
│   │
│   └── models/
│       ├── player.go            // Player model
│       ├── session.go           // Game session model
│       └── enemy.go             // Enemy model
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── .env
└── README.md
```

## Prerequisites

- Go 1.22 or higher
- PostgreSQL 16
- Redis 7

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
DATABASE_URL=postgres://gameuser:secret@postgres:5432/realtime_game?sslmode=disable
REDIS_URL=redis:6379
```

## Running the Application

### Using Docker Compose

```bash
docker-compose up
```

### Running Locally

1. Start PostgreSQL and Redis
2. Set up the environment variables
3. Run the application:

```bash
go run cmd/server/main.go
```

## WebSocket API

Connect to the WebSocket endpoint at `/ws` with optional query parameters:

```
ws://localhost:3000/ws?playerId=123&roomId=456
```

### Message Format

```json
{
  "type": "message_type",
  "payload": {},
  "roomId": "optional_room_id",
  "senderId": "optional_sender_id"
}
```

### Message Types

- `join_room`: Join a game room
- `leave_room`: Leave a game room
- `ready`: Mark player as ready
- `deal_cards`: Deal cards to a player
- `hold_card`: Hold a card for the next round
- `discard_card`: Discard a card
- `place_tower`: Place a tower
- `upgrade_tower`: Upgrade a tower
- `start_wave`: Start an enemy wave
- `game_state`: Update game state

## Game Mechanics

### Poker Hands

The game uses standard poker hand rankings:

1. Royal Flush
2. Straight Flush
3. Four of a Kind
4. Full House
5. Flush
6. Straight
7. Three of a Kind
8. Two Pair
9. Pair
10. High Card

### Tower Types

- Basic Tower: Balanced stats
- Splash Tower: Area damage
- Sniper Tower: High damage, long range
- Slow Tower: Slows enemies

### Enemy Types

- Basic: Balanced stats
- Fast: High speed, low health
- Tank: High health, low speed
- Boss: Very high health and damage

## License

MIT 