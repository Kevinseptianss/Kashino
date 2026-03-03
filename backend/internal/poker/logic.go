package poker

import (
	"fmt"
	"kashino-backend/internal/models"
	"log"
	"math/rand"
	"sort"
	"time"
)

var Suits = []string{"spades", "hearts", "diamonds", "clubs"}
var Values = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "jack", "queen", "king", "ace"}

func NewDeck() []models.Card {
	deck := make([]models.Card, 0, 52)
	for _, suit := range Suits {
		for _, value := range Values {
			deck = append(deck, models.Card{Suit: suit, Value: value})
		}
	}
	return deck
}

func Shuffle(deck []models.Card) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

var currentDeck []models.Card

type BalanceManager interface {
	UpdateBalance(userID string, amount int64, source string)
	NotifyRoomUpdate(action string, room *models.Room)
	LogPokerEvent(roomID string, handID string, event string, playerID string, username string, amount int64, pot int64, details string)
}

func StartHand(room *models.Room, bm BalanceManager) {
	room.GameState.ID = fmt.Sprintf("h_%s_%d", room.ID, time.Now().Unix())
	if room.GameState.Round != "waiting" {
		return
	}

	activePlayers := 0
	for i := range room.GameState.Players {
		if room.GameState.Players[i].Balance > 0 {
			room.GameState.Players[i].InGame = true
			activePlayers++
		} else {
			room.GameState.Players[i].InGame = false
		}
	}

	if activePlayers < 2 {
		return
	}

	room.GameState.Round = "pre-flop"
	room.GameState.Community = nil
	room.GameState.Pot = 0
	room.GameState.LastWinners = nil

	currentDeck = NewDeck()
	Shuffle(currentDeck)

	cursor := 0
	for i := range room.GameState.Players {
		if room.GameState.Players[i].InGame {
			room.GameState.Players[i].Cards = []models.Card{currentDeck[cursor], currentDeck[cursor+1]}
			cursor += 2
			room.GameState.Players[i].IsFolded = false
			room.GameState.Players[i].LastBet = 0
			room.GameState.Players[i].HasActed = false
			room.GameState.Players[i].HandContribution = 0
		}
	}

	n := len(room.GameState.Players)
	room.GameState.DealerIdx = (room.GameState.DealerIdx + 1) % n

	sbIdx := (room.GameState.DealerIdx + 1) % n
	bbIdx := (room.GameState.DealerIdx + 2) % n

	// Deduct Small Blind
	sbPlayer := &room.GameState.Players[sbIdx]
	sbPlayer.LastBet = room.SmallBlind
	sbPlayer.HandContribution = room.SmallBlind
	sbPlayer.Balance -= room.SmallBlind
	bm.UpdateBalance(sbPlayer.ID, -room.SmallBlind, "poker_blind")
	bm.LogPokerEvent(room.ID, room.GameState.ID, "small_blind", sbPlayer.ID, sbPlayer.Username, room.SmallBlind, room.SmallBlind, "Small Blind")

	// Deduct Big Blind
	bbPlayer := &room.GameState.Players[bbIdx]
	bbPlayer.LastBet = room.BigBlind
	bbPlayer.HandContribution = room.BigBlind
	bbPlayer.Balance -= room.BigBlind
	bm.UpdateBalance(bbPlayer.ID, -room.BigBlind, "poker_blind")
	bm.LogPokerEvent(room.ID, room.GameState.ID, "big_blind", bbPlayer.ID, bbPlayer.Username, room.BigBlind, room.SmallBlind+room.BigBlind, "Big Blind")

	room.GameState.Pot = room.SmallBlind + room.BigBlind
	room.GameState.CurrentTurn = room.GameState.Players[(bbIdx+1)%n].ID

	updateHandRanks(room)
}

func updateHandRanks(room *models.Room) {
	for i := range room.GameState.Players {
		p := &room.GameState.Players[i]
		if p.InGame && !p.IsFolded {
			p.CurrentHand = GetHandRank(p.Cards, room.GameState.Community)
		} else {
			p.CurrentHand = ""
		}
	}
}

func HandleAction(room *models.Room, playerID string, action models.PokerAction, bm BalanceManager) {
	if room.GameState.CurrentTurn != playerID {
		log.Printf("POKER: Action rejected. Turn is %s, but %s tried to act", room.GameState.CurrentTurn, playerID)
		return
	}
	log.Printf("POKER: Handling action '%s' (amt %d) from %s", action.Action, action.Amount, playerID)

	switch action.Action {
	case "fold":
		for i := range room.GameState.Players {
			if room.GameState.Players[i].ID == playerID {
				room.GameState.Players[i].IsFolded = true
				room.GameState.Players[i].HasActed = true
				bm.LogPokerEvent(room.ID, room.GameState.ID, "fold", playerID, room.GameState.Players[i].Username, 0, room.GameState.Pot, "Fold")
				break
			}
		}
	case "check", "call":
		// Find max bet among active, non-folded players
		var maxBet int64 = 0
		for _, p := range room.GameState.Players {
			if p.InGame && !p.IsFolded && p.LastBet > maxBet {
				maxBet = p.LastBet
			}
		}

		for i := range room.GameState.Players {
			if room.GameState.Players[i].ID == playerID {
				diff := maxBet - room.GameState.Players[i].LastBet
				if diff > 0 {
					room.GameState.Players[i].Balance -= diff
					room.GameState.Pot += diff
					room.GameState.Players[i].LastBet = maxBet
					room.GameState.Players[i].HandContribution += diff
					bm.UpdateBalance(playerID, -diff, "poker_bet")
					bm.LogPokerEvent(room.ID, room.GameState.ID, action.Action, playerID, room.GameState.Players[i].Username, diff, room.GameState.Pot, action.Action)
				} else {
					bm.LogPokerEvent(room.ID, room.GameState.ID, action.Action, playerID, room.GameState.Players[i].Username, 0, room.GameState.Pot, action.Action)
				}
				room.GameState.Players[i].HasActed = true
				break
			}
		}
	case "raise":
		for i := range room.GameState.Players {
			if room.GameState.Players[i].ID == playerID {
				diff := action.Amount - room.GameState.Players[i].LastBet
				if diff > 0 {
					room.GameState.Players[i].Balance -= diff
					room.GameState.Pot += diff
					room.GameState.Players[i].LastBet += diff
					room.GameState.Players[i].HandContribution += diff
					bm.UpdateBalance(playerID, -diff, "poker_bet")

					// When someone raises, everyone else must act again
					bm.LogPokerEvent(room.ID, room.GameState.ID, "raise", playerID, room.GameState.Players[i].Username, diff, room.GameState.Pot, "Raise to "+fmt.Sprintf("%d", action.Amount))
					for j := range room.GameState.Players {
						if room.GameState.Players[j].ID == playerID {
							room.GameState.Players[j].HasActed = true
						} else if room.GameState.Players[j].InGame && !room.GameState.Players[j].IsFolded {
							room.GameState.Players[j].HasActed = false
						}
					}
				} else {
					log.Printf("POKER: Pseudo-raise from %s (diff %d <= 0)", playerID, diff)
					room.GameState.Players[i].HasActed = true
				}
				break
			}
		}
	}

	updateHandRanks(room)
	nextTurn(room, bm)
}

func nextTurn(room *models.Room, bm BalanceManager) {
	// Simple turn progression for demo
	// Check if only one player left
	activeCount := 0
	for _, p := range room.GameState.Players {
		if !p.IsFolded && p.InGame {
			activeCount++
		}
	}
	if activeCount <= 1 {
		EndHand(room, bm)
		return
	}

	// Check if betting is settled
	var maxBet int64 = 0
	for _, p := range room.GameState.Players {
		if p.InGame && !p.IsFolded && p.LastBet > maxBet {
			maxBet = p.LastBet
		}
	}

	allActedAndMatched := true
	for _, p := range room.GameState.Players {
		if p.InGame && !p.IsFolded {
			diff := maxBet - p.LastBet
			if !p.HasActed || diff > 0 {
				log.Printf("POKER: Round NOT settled. Player %s: HasActed=%t, LastBet=%d, MaxBet=%d, Diff=%d", p.Username, p.HasActed, p.LastBet, maxBet, diff)
				allActedAndMatched = false
				break
			}
		}
	}

	if allActedAndMatched {
		log.Printf("POKER: All acted and matched. Advancing from %s", room.GameState.Round)
		advanceRound(room, bm)
		return
	}

	currIdx := -1
	for i, p := range room.GameState.Players {
		if p.ID == room.GameState.CurrentTurn {
			currIdx = i
			break
		}
	}

	n := len(room.GameState.Players)
	for i := 1; i <= n; i++ {
		nextIdx := (currIdx + i) % n
		p := &room.GameState.Players[nextIdx]
		if p.InGame && !p.IsFolded {
			diff := maxBet - p.LastBet
			if !p.HasActed || diff > 0 {
				log.Printf("POKER: Next turn assigned to %s (Idx %d). HasActed=%t, Diff=%d", p.Username, nextIdx, p.HasActed, diff)
				room.GameState.CurrentTurn = p.ID
				return
			}
		}
	}

	// Safety fallback: if we can't find anyone, advance
	advanceRound(room, bm)
}

func advanceRound(room *models.Room, bm BalanceManager) {
	switch room.GameState.Round {
	case "pre-flop":
		room.GameState.Round = "flop"
		room.GameState.Community = append(room.GameState.Community, currentDeck[20:23]...)
	case "flop":
		room.GameState.Round = "turn"
		room.GameState.Community = append(room.GameState.Community, currentDeck[23])
	case "turn":
		room.GameState.Round = "river"
		room.GameState.Community = append(room.GameState.Community, currentDeck[24])
	case "river":
		log.Printf("POKER: Showdown! Pot: %d", room.GameState.Pot)
		EndHand(room, bm)
		return
	}
	log.Printf("POKER: Round advanced to %s. Pot: %d, Community: %d", room.GameState.Round, room.GameState.Pot, len(room.GameState.Community))

	updateHandRanks(room)

	// Reset bets for the new round
	for i := range room.GameState.Players {
		room.GameState.Players[i].LastBet = 0
		room.GameState.Players[i].HasActed = false
	}

	// Reset current turn to first active player after dealer
	n := len(room.GameState.Players)
	for i := 1; i <= n; i++ {
		idx := (room.GameState.DealerIdx + i) % n
		if !room.GameState.Players[idx].IsFolded && room.GameState.Players[idx].InGame {
			room.GameState.CurrentTurn = room.GameState.Players[idx].ID
			break
		}
	}
}

func EndHand(room *models.Room, bm BalanceManager) {
	log.Printf("POKER: Entering EndHand. Total Pot: %d", room.GameState.Pot)
	room.GameState.LastWinners = nil

	// 1. Calculate side pots
	// Unique contributions from all players (including folded ones)
	contribMap := make(map[int64]bool)
	for _, p := range room.GameState.Players {
		if p.HandContribution > 0 {
			contribMap[p.HandContribution] = true
		}
	}

	var thresholds []int64
	for t := range contribMap {
		thresholds = append(thresholds, t)
	}
	sort.Slice(thresholds, func(i, j int) bool { return thresholds[i] < thresholds[j] })

	prevThreshold := int64(0)
	for _, t := range thresholds {
		potSize := int64(0)
		rangeDiff := t - prevThreshold
		var eligibleWinners []*models.Player

		for i := range room.GameState.Players {
			p := &room.GameState.Players[i]
			// Money contribution to this specific sub-pot
			if p.HandContribution >= t {
				potSize += rangeDiff
				if p.InGame && !p.IsFolded {
					eligibleWinners = append(eligibleWinners, p)
				}
			} else if p.HandContribution > prevThreshold {
				potSize += (p.HandContribution - prevThreshold)
			}
		}

		if potSize > 0 && len(eligibleWinners) > 0 {
			// Find high score for this sub-pot
			maxScore := -1
			for _, p := range eligibleWinners {
				score := GetHandScore(p.Cards, room.GameState.Community)
				if score > maxScore {
					maxScore = score
				}
			}

			// Find winners for this sub-pot
			var potWinners []*models.Player
			for _, p := range eligibleWinners {
				score := GetHandScore(p.Cards, room.GameState.Community)
				if score == maxScore {
					potWinners = append(potWinners, p)
				}
			}

			// Distribute sub-pot
			share := potSize / int64(len(potWinners))
			remainder := potSize % int64(len(potWinners))

			for i, w := range potWinners {
				amount := share
				if i == 0 {
					amount += remainder // Small remainder goes to first winner
				}
				w.Balance += amount
				bm.UpdateBalance(w.ID, amount, "poker_win")
				bm.LogPokerEvent(room.ID, room.GameState.ID, "win", w.ID, w.Username, amount, room.GameState.Pot, "Sub-pot Winner")

				// Update LastWinners for UI reveal
				found := false
				for j := range room.GameState.LastWinners {
					if room.GameState.LastWinners[j].UserID == w.ID {
						room.GameState.LastWinners[j].Amount += amount
						found = true
						break
					}
				}
				if !found {
					room.GameState.LastWinners = append(room.GameState.LastWinners, models.WinnerInfo{
						UserID:          w.ID,
						Username:        w.Username,
						Amount:          amount,
						HandDescription: w.CurrentHand,
					})
				}
			}
		} else if potSize > 0 {
			log.Printf("POKER: WARNING: Dead sub-pot of %d with no eligible winners", potSize)
			// Emergency: Give it to the house or leave in pot?
			// In normally logic, someone is ALWAYS eligible for the first sub-pot.
		}

		prevThreshold = t
	}

	room.GameState.Round = "waiting"
	room.GameState.CurrentTurn = ""
	room.GameState.Pot = 0

	// Clear cards for next hand
	for i := range room.GameState.Players {
		room.GameState.Players[i].Cards = nil
		room.GameState.Players[i].LastBet = 0
		room.GameState.Players[i].HandContribution = 0
		room.GameState.Players[i].CurrentHand = ""
	}
}

func LeaveGame(room *models.Room, playerID string, bm BalanceManager) {
	idx := -1
	var username string
	for i, p := range room.GameState.Players {
		if p.ID == playerID {
			idx = i
			username = p.Username
			break
		}
	}

	if idx == -1 {
		return
	}

	isHandActive := room.GameState.Round != "waiting"
	isTurn := room.GameState.CurrentTurn == playerID

	if isHandActive {
		bm.LogPokerEvent(room.ID, room.GameState.ID, "leave", playerID, username, 0, room.GameState.Pot, "Player Left/Stood Up")
		// Force fold if in hand
		room.GameState.Players[idx].IsFolded = true
		room.GameState.Players[idx].HasActed = true
		room.GameState.Players[idx].InGame = false

		if isTurn {
			nextTurn(room, bm)
		}
	}

	// Adjust DealerIdx
	// If the removed player's index is <= DealerIdx, we need to shift DealerIdx
	if idx <= room.GameState.DealerIdx && room.GameState.DealerIdx > 0 {
		room.GameState.DealerIdx--
	} else if idx <= room.GameState.DealerIdx && room.GameState.DealerIdx == 0 {
		// If dealer was at 0 and we remove 0, dealer moves to the new end of slice (handled by modulo usually)
		if len(room.GameState.Players) > 1 {
			room.GameState.DealerIdx = len(room.GameState.Players) - 2
			if room.GameState.DealerIdx < 0 {
				room.GameState.DealerIdx = 0
			}
		} else {
			room.GameState.DealerIdx = 0
		}
	}

	// Remove from slice
	room.GameState.Players = append(room.GameState.Players[:idx], room.GameState.Players[idx+1:]...)

	// Check if hand should end
	if isHandActive {
		activeCount := 0
		for _, p := range room.GameState.Players {
			if !p.IsFolded && p.InGame {
				activeCount++
			}
		}
		if activeCount <= 1 {
			EndHand(room, bm)
		}
	}
}
