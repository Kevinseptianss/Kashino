package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Card struct {
	Suit  string `json:"suit"`  // "spades", "hearts", "diamonds", "clubs"
	Value string `json:"value"` // "2", "3", ..., "10", "jack", "queen", "king", "ace"
}

type Player struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Balance     float64 `json:"balance"`
	Cards       []Card  `json:"cards,omitempty"`
	Position    int     `json:"position"` // 0-4
	LastBet     float64 `json:"last_bet"`
	IsFolded    bool    `json:"is_folded"`
	InGame      bool    `json:"in_game"` // Sitting vs Playing
	CurrentHand string  `json:"current_hand,omitempty"`
}

type WinnerInfo struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	Amount          float64 `json:"amount"`
	HandDescription string  `json:"hand_description"`
}

type GameState struct {
	ID          string       `json:"id"`
	Players     []Player     `json:"players"`
	Community   []Card       `json:"community"`
	Pot         float64      `json:"pot"`
	CurrentTurn string       `json:"current_turn"`
	Round       string       `json:"round"` // "waiting", "pre-flop", "flop", "turn", "river", "showdown"
	DealerIdx   int          `json:"dealer_idx"`
	LastWinners []WinnerInfo `json:"last_winners,omitempty"`
}

type Room struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	GameState  GameState `json:"game_state"`
	MaxPlayers int       `json:"max_players"`
	SmallBlind float64   `json:"small_blind"`
	BigBlind   float64   `json:"big_blind"`
}

type PokerAction struct {
	Action string  `json:"action"` // "sit", "check", "call", "raise", "fold"
	Amount float64 `json:"amount,omitempty"`
	Seat   int     `json:"seat,omitempty"`
}

type PokerHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    string             `bson:"room_id" json:"room_id"`
	HandID    string             `bson:"hand_id" json:"hand_id"`
	Event     string             `bson:"event" json:"event"` // "start", "action", "win", "fold"
	PlayerID  string             `bson:"player_id,omitempty" json:"player_id,omitempty"`
	Username  string             `bson:"username,omitempty" json:"username,omitempty"`
	Amount    float64            `bson:"amount,omitempty" json:"amount,omitempty"`
	Pot       float64            `bson:"pot" json:"pot"`
	Cards     []Card             `bson:"cards,omitempty" json:"cards,omitempty"`
	Community []Card             `bson:"community,omitempty" json:"community,omitempty"`
	Details   string             `bson:"details,omitempty" json:"details,omitempty"`
	Timestamp primitive.DateTime `bson:"timestamp" json:"timestamp"`
}
