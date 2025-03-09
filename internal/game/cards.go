package game

import (
	"math/rand"
	"time"

	"realtime-game-backend/internal/models"
)

// Suits and ranks for a standard deck of cards
var (
	suits  = []string{"hearts", "diamonds", "clubs", "spades"}
	ranks  = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	values = map[string]int{
		"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9, "10": 10,
		"J": 11, "Q": 12, "K": 13, "A": 14,
	}
)

// NewDeck creates a new deck of cards
func NewDeck() []models.Card {
	var deck []models.Card

	for _, suit := range suits {
		for _, rank := range ranks {
			card := models.Card{
				ID:     suit + "-" + rank,
				Suit:   suit,
				Rank:   rank,
				Value:  values[rank],
				Held:   false,
				Active: true,
			}
			deck = append(deck, card)
		}
	}

	return deck
}

// ShuffleDeck shuffles a deck of cards
func ShuffleDeck(deck []models.Card) []models.Card {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Fisher-Yates shuffle algorithm
	for i := len(deck) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	return deck
}

// DealCards deals a specified number of cards from a deck
func DealCards(deck []models.Card, count int) ([]models.Card, []models.Card) {
	if count > len(deck) {
		count = len(deck)
	}

	hand := deck[:count]
	remainingDeck := deck[count:]

	return hand, remainingDeck
}

// DrawCards draws new cards to replace discarded ones
func DrawCards(hand []models.Card, deck []models.Card) ([]models.Card, []models.Card) {
	var newHand []models.Card

	// Keep held cards
	for _, card := range hand {
		if card.Held {
			newHand = append(newHand, card)
		}
	}

	// Draw new cards to replace discarded ones
	cardsNeeded := 5 - len(newHand)
	if cardsNeeded > 0 {
		drawnCards, remainingDeck := DealCards(deck, cardsNeeded)
		newHand = append(newHand, drawnCards...)
		deck = remainingDeck
	}

	// Reset held status for next round
	for i := range newHand {
		newHand[i].Held = false
	}

	return newHand, deck
}

// HoldCard marks a card as held for the next round
func HoldCard(hand []models.Card, cardID string) []models.Card {
	for i, card := range hand {
		if card.ID == cardID {
			hand[i].Held = true
			break
		}
	}

	return hand
}

// DiscardCard marks a card as not held for the next round
func DiscardCard(hand []models.Card, cardID string) []models.Card {
	for i, card := range hand {
		if card.ID == cardID {
			hand[i].Held = false
			break
		}
	}

	return hand
}

// GetCardByID gets a card by its ID
func GetCardByID(cards []models.Card, cardID string) *models.Card {
	for i, card := range cards {
		if card.ID == cardID {
			return &cards[i]
		}
	}

	return nil
}
