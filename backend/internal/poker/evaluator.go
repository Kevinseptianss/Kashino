package poker

import (
	"kashino-backend/internal/models"
	"sort"
)

func GetHandRank(holeCards []models.Card, communityCards []models.Card) string {
	allCards := append([]models.Card{}, holeCards...)
	allCards = append(allCards, communityCards...)

	if len(allCards) < 2 {
		return ""
	}

	// Sort cards by value
	sort.Slice(allCards, func(i, j int) bool {
		iv := cardValueToInt(allCards[i].Value)
		jv := cardValueToInt(allCards[j].Value)
		if iv == jv {
			return allCards[i].Suit > allCards[j].Suit
		}
		return iv > jv
	})

	if isFlush(allCards) {
		if isStraight(allCards) {
			return "Straight Flush"
		}
		return "Flush"
	}

	if isStraight(allCards) {
		return "Straight"
	}

	counts := getCounts(allCards)
	if hasCount(counts, 4) {
		return "Four of a Kind"
	}
	if hasCount(counts, 3) && hasCount(counts, 2) {
		return "Full House"
	}
	if hasCount(counts, 3) {
		return "Three of a Kind"
	}
	pairCount := 0
	for _, count := range counts {
		if count == 2 {
			pairCount++
		}
	}
	if pairCount >= 2 {
		return "Two Pair"
	}
	if pairCount == 1 {
		return "One Pair"
	}

	return "High Card: " + allCards[0].Value
}

func GetHandScore(holeCards []models.Card, communityCards []models.Card) int {
	allCards := append([]models.Card{}, holeCards...)
	allCards = append(allCards, communityCards...)

	if len(allCards) < 2 {
		return 0
	}

	// Sort cards by value
	sort.Slice(allCards, func(i, j int) bool {
		iv := cardValueToInt(allCards[i].Value)
		jv := cardValueToInt(allCards[j].Value)
		return iv > jv
	})

	// Basic scoring: Rank * 1000000 + TopCardValue * 10000 + Kickers...
	// Simplified for now: just Rank and Top Card
	rank := 0
	if isFlush(allCards) {
		if isStraight(allCards) {
			rank = 8
		} else {
			rank = 5
		}
	} else if isStraight(allCards) {
		rank = 4
	} else {
		counts := getCounts(allCards)
		if hasCount(counts, 4) {
			rank = 7
		} else if hasCount(counts, 3) && hasCount(counts, 2) {
			rank = 6
		} else if hasCount(counts, 3) {
			rank = 3
		} else {
			pairCount := 0
			for _, count := range counts {
				if count == 2 {
					pairCount++
				}
			}
			if pairCount >= 2 {
				rank = 2
			} else if pairCount == 1 {
				rank = 1
			}
		}
	}

	topValue := cardValueToInt(allCards[0].Value)
	return rank*100 + topValue
}

func cardValueToInt(val string) int {
	switch val {
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	case "5":
		return 5
	case "6":
		return 6
	case "7":
		return 7
	case "8":
		return 8
	case "9":
		return 9
	case "10":
		return 10
	case "jack":
		return 11
	case "queen":
		return 12
	case "king":
		return 13
	case "ace":
		return 14
	}
	return 0
}

func getCounts(cards []models.Card) map[string]int {
	counts := make(map[string]int)
	for _, c := range cards {
		counts[c.Value]++
	}
	return counts
}

func hasCount(counts map[string]int, target int) bool {
	for _, c := range counts {
		if c == target {
			return true
		}
	}
	return false
}

func isFlush(cards []models.Card) bool {
	suits := make(map[string]int)
	for _, c := range cards {
		suits[c.Suit]++
		if suits[c.Suit] >= 5 {
			return true
		}
	}
	return false
}

func isStraight(cards []models.Card) bool {
	// Remove duplicate values for straight check
	uniqueValues := make([]int, 0)
	seen := make(map[int]bool)
	for _, c := range cards {
		v := cardValueToInt(c.Value)
		if !seen[v] {
			uniqueValues = append(uniqueValues, v)
			seen[v] = true
		}
	}
	sort.Ints(uniqueValues)

	// Check for Ace-low straight (A-2-3-4-5)
	if seen[14] && seen[2] && seen[3] && seen[4] && seen[5] {
		return true
	}

	count := 1
	for i := 0; i < len(uniqueValues)-1; i++ {
		if uniqueValues[i+1] == uniqueValues[i]+1 {
			count++
			if count >= 5 {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}
