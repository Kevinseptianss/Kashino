package models

type Card struct {
	Suit  string `json:"suit"`
	Value string `json:"value"`
}

type Player struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
	Cards    []Card  `json:"cards,omitempty"`
	IsDealer bool    `json:"is_dealer"`
	LastBet  float64 `json:"last_bet"`
}

type GameState struct {
	ID          string   `json:"id"`
	Players     []Player `json:"players"`
	Community   []Card   `json:"community"`
	Pot         float64  `json:"pot"`
	CurrentTurn string   `json:"current_turn"`
	Round       string   `json:"round"` // e.g., "pre-flop", "flop", "turn", "river"
}
