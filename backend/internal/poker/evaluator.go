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

	// Sort cards by value descending
	sort.Slice(allCards, func(i, j int) bool {
		iv := cardValueToInt(allCards[i].Value)
		jv := cardValueToInt(allCards[j].Value)
		if iv == jv {
			return allCards[i].Suit > allCards[j].Suit
		}
		return iv > jv
	})

	counts := getCounts(allCards)

	if sf, _ := findStraightFlush(allCards); sf {
		return "Straight Flush"
	}
	if hasSet, _ := findNOfAKind(counts, 4, -1); hasSet {
		return "Four of a Kind"
	}
	if hasSet, v := findNOfAKind(counts, 3, -1); hasSet {
		if hasPair, _ := findNOfAKind(counts, 2, v); hasPair {
			return "Full House"
		}
		if hasSecondSet, _ := findNOfAKind(counts, 3, v); hasSecondSet {
			return "Full House"
		}
	}
	if f, _ := findFlush(allCards); f {
		return "Flush"
	}
	if s, _ := findStraight(allCards); s {
		return "Straight"
	}
	if hasSet, _ := findNOfAKind(counts, 3, -1); hasSet {
		return "Three of a Kind"
	}
	if hasPair, v := findNOfAKind(counts, 2, -1); hasPair {
		if hasSecondPair, _ := findNOfAKind(counts, 2, v); hasSecondPair {
			return "Two Pair"
		}
		return "One Pair"
	}

	return "High Card: " + allCards[0].Value
}

func GetHandScore(holeCards []models.Card, communityCards []models.Card) int {
	allCards := append([]models.Card{}, holeCards...)
	allCards = append(allCards, communityCards...)

	if len(allCards) < 5 {
		// Minimum 5 cards are needed for a full evaluation in Hold'em
		// but we can still give a basic score for fewer cards
		sort.Slice(allCards, func(i, j int) bool {
			return cardValueToInt(allCards[i].Value) > cardValueToInt(allCards[j].Value)
		})
		score := 0
		for i := 0; i < len(allCards) && i < 5; i++ {
			score |= cardValueToInt(allCards[i].Value) << (4 * (4 - i))
		}
		return score
	}

	// Sort cards by value descending
	sort.Slice(allCards, func(i, j int) bool {
		iv := cardValueToInt(allCards[i].Value)
		jv := cardValueToInt(allCards[j].Value)
		if iv == jv {
			return allCards[i].Suit > allCards[j].Suit
		}
		return iv > jv
	})

	// 1. Straight Flush
	sf, high := findStraightFlush(allCards)
	if sf {
		return (8 << 20) | (high << 16)
	}

	// 2. Four of a Kind
	counts := getCounts(allCards)
	for val, count := range counts {
		if count == 4 {
			v := cardValueToInt(val)
			kicker := 0
			for _, c := range allCards {
				cv := cardValueToInt(c.Value)
				if cv != v {
					kicker = cv
					break
				}
			}
			return (7 << 20) | (v << 16) | (kicker << 12)
		}
	}

	// 3. Full House
	hasSet, setVal := findNOfAKind(counts, 3, -1)
	if hasSet {
		hasPair, pairVal := findNOfAKind(counts, 2, -1)
		if hasPair {
			return (6 << 20) | (setVal << 16) | (pairVal << 12)
		}
		// Check for a second set (which counts as a full house pair)
		hasSecondSet, secondSetVal := findNOfAKind(counts, 3, setVal)
		if hasSecondSet {
			return (6 << 20) | (setVal << 16) | (secondSetVal << 12)
		}
	}

	// 4. Flush
	f, flushCards := findFlush(allCards)
	if f {
		score := (5 << 20)
		for i := 0; i < 5; i++ {
			score |= cardValueToInt(flushCards[i].Value) << (4 * (4 - i))
		}
		return score
	}

	// 5. Straight
	s, sHigh := findStraight(allCards)
	if s {
		return (4 << 20) | (sHigh << 16)
	}

	// 6. Three of a Kind
	if hasSet {
		score := (3 << 20) | (setVal << 16)
		kIdx := 0
		for _, c := range allCards {
			cv := cardValueToInt(c.Value)
			if cv != setVal && kIdx < 2 {
				score |= cv << (4 * (2 - kIdx))
				kIdx++
			}
		}
		return score
	}

	// 7. Two Pair
	hasPair1, pair1Val := findNOfAKind(counts, 2, -1)
	if hasPair1 {
		hasPair2, pair2Val := findNOfAKind(counts, 2, pair1Val)
		if hasPair2 {
			kicker := 0
			for _, c := range allCards {
				cv := cardValueToInt(c.Value)
				if cv != pair1Val && cv != pair2Val {
					kicker = cv
					break
				}
			}
			return (2 << 20) | (pair1Val << 16) | (pair2Val << 12) | (kicker << 8)
		}
	}

	// 8. One Pair
	if hasPair1 {
		score := (1 << 20) | (pair1Val << 16)
		kIdx := 0
		for _, c := range allCards {
			cv := cardValueToInt(c.Value)
			if cv != pair1Val && kIdx < 3 {
				score |= cv << (4 * (2 - kIdx))
				kIdx++
			}
		}
		return score
	}

	// 9. High Card
	score := 0
	for i := 0; i < 5; i++ {
		score |= cardValueToInt(allCards[i].Value) << (4 * (4 - i))
	}
	return score
}

func findStraightFlush(cards []models.Card) (bool, int) {
	suits := make(map[string][]models.Card)
	for _, c := range cards {
		suits[c.Suit] = append(suits[c.Suit], c)
	}
	for _, suitCards := range suits {
		if len(suitCards) >= 5 {
			s, high := findStraight(suitCards)
			if s {
				return true, high
			}
		}
	}
	return false, 0
}

func findFlush(cards []models.Card) (bool, []models.Card) {
	suits := make(map[string][]models.Card)
	for _, c := range cards {
		suits[c.Suit] = append(suits[c.Suit], c)
	}
	for _, suitCards := range suits {
		if len(suitCards) >= 5 {
			return true, suitCards[:5]
		}
	}
	return false, nil
}

func findStraight(cards []models.Card) (bool, int) {
	seen := make(map[int]bool)
	unique := []int{}
	for _, c := range cards {
		v := cardValueToInt(c.Value)
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	sort.Slice(unique, func(i, j int) bool { return unique[i] > unique[j] })

	if len(unique) < 5 {
		return false, 0
	}

	// A-2-3-4-5
	if seen[14] && seen[2] && seen[3] && seen[4] && seen[5] {
		// Check for higher straights first, but if this is the only one...
		// We'll return 5 as the high card for this specific case at the end if no others found.
	}

	for i := 0; i <= len(unique)-5; i++ {
		if unique[i] == unique[i+4]+4 {
			return true, unique[i]
		}
	}

	if seen[14] && seen[2] && seen[3] && seen[4] && seen[5] {
		return true, 5
	}

	return false, 0
}

func findNOfAKind(counts map[string]int, n int, exclude int) (bool, int) {
	maxVal := -1
	for val, count := range counts {
		v := cardValueToInt(val)
		if count >= n && v != exclude {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	return maxVal != -1, maxVal
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
