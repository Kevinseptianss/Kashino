package poker

import (
	"fmt"
	"kashino-backend/internal/models"
	"math/rand"
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
	UpdateBalance(userID string, amount float64, source string)
	NotifyRoomUpdate(action string, room *models.Room)
	LogPokerEvent(roomID string, handID string, event string, playerID string, username string, amount float64, pot float64, details string)
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
		}
	}

	n := len(room.GameState.Players)
	room.GameState.DealerIdx = (room.GameState.DealerIdx + 1) % n

	sbIdx := (room.GameState.DealerIdx + 1) % n
	bbIdx := (room.GameState.DealerIdx + 2) % n

	// Deduct Small Blind
	sbPlayer := &room.GameState.Players[sbIdx]
	sbPlayer.LastBet = room.SmallBlind
	sbPlayer.Balance -= room.SmallBlind
	bm.UpdateBalance(sbPlayer.ID, -room.SmallBlind, "poker_blind")
	bm.LogPokerEvent(room.ID, room.GameState.ID, "small_blind", sbPlayer.ID, sbPlayer.Username, room.SmallBlind, room.SmallBlind, "Small Blind")

	// Deduct Big Blind
	bbPlayer := &room.GameState.Players[bbIdx]
	bbPlayer.LastBet = room.BigBlind
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
		return
	}

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
		// Find max bet at table
		maxBet := 0.0
		for _, p := range room.GameState.Players {
			if p.LastBet > maxBet {
				maxBet = p.LastBet
			}
		}

		for i := range room.GameState.Players {
			if room.GameState.Players[i].ID == playerID {
				diff := maxBet - room.GameState.Players[i].LastBet
				if diff > 0 {
					room.GameState.Players[i].Balance -= diff
					room.GameState.Pot += diff
					room.GameState.Players[i].LastBet += diff
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
					bm.UpdateBalance(playerID, -diff, "poker_bet")
					bm.LogPokerEvent(room.ID, room.GameState.ID, "raise", playerID, room.GameState.Players[i].Username, diff, room.GameState.Pot, "Raise to "+fmt.Sprintf("%.0f", action.Amount))
				}

				// When someone raises, everyone else must act again
				for j := range room.GameState.Players {
					if room.GameState.Players[j].ID == playerID {
						room.GameState.Players[j].HasActed = true
					} else if room.GameState.Players[j].InGame && !room.GameState.Players[j].IsFolded {
						room.GameState.Players[j].HasActed = false
					}
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
	maxBet := 0.0
	for _, p := range room.GameState.Players {
		if p.InGame && !p.IsFolded && p.LastBet > maxBet {
			maxBet = p.LastBet
		}
	}

	allActedAndMatched := true
	for _, p := range room.GameState.Players {
		if p.InGame && !p.IsFolded {
			if !p.HasActed || p.LastBet != maxBet {
				allActedAndMatched = false
				break
			}
		}
	}

	if allActedAndMatched {
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
			// If betting is not settled, and this player hasn't matched or hasn't acted, it's their turn
			if !p.HasActed || p.LastBet != maxBet {
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
		EndHand(room, bm)
		return
	}

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
	// 1. Identify all active, non-folded players
	var eligiblePlayers []*models.Player
	for i := range room.GameState.Players {
		if !room.GameState.Players[i].IsFolded && room.GameState.Players[i].InGame {
			eligiblePlayers = append(eligiblePlayers, &room.GameState.Players[i])
		}
	}

	// 2. Find the highest score among eligible players
	maxScore := -1
	for _, p := range eligiblePlayers {
		score := GetHandScore(p.Cards, room.GameState.Community)
		if score > maxScore {
			maxScore = score
		}
	}

	// 3. Find all winners who share the highest score
	var winners []*models.Player
	for _, p := range eligiblePlayers {
		score := GetHandScore(p.Cards, room.GameState.Community)
		if score == maxScore {
			winners = append(winners, p)
		}
	}

	room.GameState.LastWinners = nil
	if len(winners) > 0 {
		// Divide pot among winners (handle split pot)
		share := room.GameState.Pot / float64(len(winners))
		for _, w := range winners {
			w.Balance += share
			bm.UpdateBalance(w.ID, share, "poker_win")
			bm.LogPokerEvent(room.ID, room.GameState.ID, "win", w.ID, w.Username, share, room.GameState.Pot, "Winner")

			room.GameState.LastWinners = append(room.GameState.LastWinners, models.WinnerInfo{
				UserID:          w.ID,
				Username:        w.Username,
				Amount:          share,
				HandDescription: w.CurrentHand,
			})
		}
	}

	room.GameState.Round = "waiting"
	room.GameState.CurrentTurn = ""
	room.GameState.Pot = 0

	// Clear cards for next hand
	for i := range room.GameState.Players {
		room.GameState.Players[i].Cards = nil
		room.GameState.Players[i].LastBet = 0
		room.GameState.Players[i].CurrentHand = ""
	}
}
