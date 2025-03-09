package game

import (
	"sort"

	"realtime-game-backend/internal/models"
)

// Hand ranks in ascending order of value
const (
	HighCard      = "high_card"
	Pair          = "pair"
	TwoPair       = "two_pair"
	ThreeOfAKind  = "three_of_a_kind"
	Straight      = "straight"
	Flush         = "flush"
	FullHouse     = "full_house"
	FourOfAKind   = "four_of_a_kind"
	StraightFlush = "straight_flush"
	RoyalFlush    = "royal_flush"
)

// Hand rank values
var handRankValues = map[string]int{
	HighCard:      1,
	Pair:          2,
	TwoPair:       3,
	ThreeOfAKind:  4,
	Straight:      5,
	Flush:         6,
	FullHouse:     7,
	FourOfAKind:   8,
	StraightFlush: 9,
	RoyalFlush:    10,
}

// Hand rank names
var handRankNames = map[string]string{
	HighCard:      "High Card",
	Pair:          "Pair",
	TwoPair:       "Two Pair",
	ThreeOfAKind:  "Three of a Kind",
	Straight:      "Straight",
	Flush:         "Flush",
	FullHouse:     "Full House",
	FourOfAKind:   "Four of a Kind",
	StraightFlush: "Straight Flush",
	RoyalFlush:    "Royal Flush",
}

// EvaluateHand evaluates a poker hand and returns its rank
func EvaluateHand(cards []models.Card) models.HandRank {
	if len(cards) != 5 {
		return models.HandRank{
			Type:  HighCard,
			Value: handRankValues[HighCard],
			Name:  handRankNames[HighCard],
		}
	}

	// Sort cards by value in descending order
	sortedCards := make([]models.Card, len(cards))
	copy(sortedCards, cards)
	sort.Slice(sortedCards, func(i, j int) bool {
		return sortedCards[i].Value > sortedCards[j].Value
	})

	// Check for royal flush
	if isRoyalFlush(sortedCards) {
		return models.HandRank{
			Type:  RoyalFlush,
			Value: handRankValues[RoyalFlush],
			Name:  handRankNames[RoyalFlush],
		}
	}

	// Check for straight flush
	if isStraightFlush(sortedCards) {
		return models.HandRank{
			Type:  StraightFlush,
			Value: handRankValues[StraightFlush],
			Name:  handRankNames[StraightFlush],
		}
	}

	// Check for four of a kind
	if isFourOfAKind(sortedCards) {
		return models.HandRank{
			Type:  FourOfAKind,
			Value: handRankValues[FourOfAKind],
			Name:  handRankNames[FourOfAKind],
		}
	}

	// Check for full house
	if isFullHouse(sortedCards) {
		return models.HandRank{
			Type:  FullHouse,
			Value: handRankValues[FullHouse],
			Name:  handRankNames[FullHouse],
		}
	}

	// Check for flush
	if isFlush(sortedCards) {
		return models.HandRank{
			Type:  Flush,
			Value: handRankValues[Flush],
			Name:  handRankNames[Flush],
		}
	}

	// Check for straight
	if isStraight(sortedCards) {
		return models.HandRank{
			Type:  Straight,
			Value: handRankValues[Straight],
			Name:  handRankNames[Straight],
		}
	}

	// Check for three of a kind
	if isThreeOfAKind(sortedCards) {
		return models.HandRank{
			Type:  ThreeOfAKind,
			Value: handRankValues[ThreeOfAKind],
			Name:  handRankNames[ThreeOfAKind],
		}
	}

	// Check for two pair
	if isTwoPair(sortedCards) {
		return models.HandRank{
			Type:  TwoPair,
			Value: handRankValues[TwoPair],
			Name:  handRankNames[TwoPair],
		}
	}

	// Check for pair
	if isPair(sortedCards) {
		return models.HandRank{
			Type:  Pair,
			Value: handRankValues[Pair],
			Name:  handRankNames[Pair],
		}
	}

	// High card
	return models.HandRank{
		Type:  HighCard,
		Value: handRankValues[HighCard],
		Name:  handRankNames[HighCard],
	}
}

// isRoyalFlush checks if the hand is a royal flush
func isRoyalFlush(cards []models.Card) bool {
	if !isFlush(cards) {
		return false
	}

	// Check if the cards are A, K, Q, J, 10 of the same suit
	values := []int{14, 13, 12, 11, 10}
	for i, value := range values {
		if cards[i].Value != value {
			return false
		}
	}

	return true
}

// isStraightFlush checks if the hand is a straight flush
func isStraightFlush(cards []models.Card) bool {
	return isFlush(cards) && isStraight(cards)
}

// isFourOfAKind checks if the hand is four of a kind
func isFourOfAKind(cards []models.Card) bool {
	// Check if the first 4 cards have the same value
	if cards[0].Value == cards[1].Value && cards[1].Value == cards[2].Value && cards[2].Value == cards[3].Value {
		return true
	}

	// Check if the last 4 cards have the same value
	if cards[1].Value == cards[2].Value && cards[2].Value == cards[3].Value && cards[3].Value == cards[4].Value {
		return true
	}

	return false
}

// isFullHouse checks if the hand is a full house
func isFullHouse(cards []models.Card) bool {
	// Check if the first 3 cards have the same value and the last 2 cards have the same value
	if cards[0].Value == cards[1].Value && cards[1].Value == cards[2].Value && cards[3].Value == cards[4].Value {
		return true
	}

	// Check if the first 2 cards have the same value and the last 3 cards have the same value
	if cards[0].Value == cards[1].Value && cards[2].Value == cards[3].Value && cards[3].Value == cards[4].Value {
		return true
	}

	return false
}

// isFlush checks if the hand is a flush
func isFlush(cards []models.Card) bool {
	suit := cards[0].Suit
	for _, card := range cards {
		if card.Suit != suit {
			return false
		}
	}
	return true
}

// isStraight checks if the hand is a straight
func isStraight(cards []models.Card) bool {
	// Special case: A-5-4-3-2
	if cards[0].Value == 14 && cards[1].Value == 5 && cards[2].Value == 4 && cards[3].Value == 3 && cards[4].Value == 2 {
		return true
	}

	// Check if the cards are in sequence
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].Value != cards[i+1].Value+1 {
			return false
		}
	}

	return true
}

// isThreeOfAKind checks if the hand is three of a kind
func isThreeOfAKind(cards []models.Card) bool {
	// Check if the first 3 cards have the same value
	if cards[0].Value == cards[1].Value && cards[1].Value == cards[2].Value {
		return true
	}

	// Check if the middle 3 cards have the same value
	if cards[1].Value == cards[2].Value && cards[2].Value == cards[3].Value {
		return true
	}

	// Check if the last 3 cards have the same value
	if cards[2].Value == cards[3].Value && cards[3].Value == cards[4].Value {
		return true
	}

	return false
}

// isTwoPair checks if the hand is two pair
func isTwoPair(cards []models.Card) bool {
	pairCount := 0
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].Value == cards[i+1].Value {
			pairCount++
			i++ // Skip the next card since it's part of the pair
		}
	}
	return pairCount == 2
}

// isPair checks if the hand is a pair
func isPair(cards []models.Card) bool {
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].Value == cards[i+1].Value {
			return true
		}
	}
	return false
}

// CompareHands compares two poker hands and returns 1 if hand1 is better, -1 if hand2 is better, and 0 if they are equal
func CompareHands(hand1, hand2 models.PokerHand) int {
	// Compare hand ranks
	if hand1.Rank.Value > hand2.Rank.Value {
		return 1
	}
	if hand1.Rank.Value < hand2.Rank.Value {
		return -1
	}

	// If the ranks are the same, compare the high cards
	// This is a simplified version that doesn't handle all tie-breaking scenarios
	sortHand := func(cards []models.Card) {
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].Value > cards[j].Value
		})
	}

	sortHand(hand1.Cards)
	sortHand(hand2.Cards)

	for i := 0; i < len(hand1.Cards); i++ {
		if hand1.Cards[i].Value > hand2.Cards[i].Value {
			return 1
		}
		if hand1.Cards[i].Value < hand2.Cards[i].Value {
			return -1
		}
	}

	return 0
}
