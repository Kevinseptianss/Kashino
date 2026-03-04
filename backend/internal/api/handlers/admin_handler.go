package handlers

import (
	"encoding/json"
	"kashino-backend/internal/models"
	"kashino-backend/internal/repository"
	"kashino-backend/internal/websocket"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminHandler struct {
	userRepo *repository.UserRepository
	hub      *websocket.Hub
}

func NewAdminHandler(userRepo *repository.UserRepository, hub *websocket.Hub) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
		hub:      hub,
	}
}

func (h *AdminHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userRepo.GetDailyStats(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	stats["online_users"] = h.hub.GetOnlineCount()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func calculateUserStats(user *models.User) models.UserStats {
	var stats models.UserStats
	var wins, games int
	for _, tx := range user.BalanceHistory {
		switch tx.Source {
		case "slot_win", "poker_win":
			stats.TotalWon += tx.Amount
			wins++
			games++
		case "slot_spin", "poker_bet", "poker_blind":
			// Amount is negative for these sources
			stats.TotalLost += -tx.Amount
			games++
		case "initial", "admin_adjustment":
			if tx.Amount > 0 {
				stats.TotalRefilled += tx.Amount
			}
		}
	}

	if games > 0 {
		stats.WinRate = (float64(wins) / float64(games)) * 100
	}
	stats.TotalGames = games
	return stats
}

func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	type UserWithStats struct {
		models.User
		Stats models.UserStats `json:"stats"`
	}

	response := make([]UserWithStats, len(users))
	for i, user := range users {
		response[i] = UserWithStats{
			User:  user,
			Stats: calculateUserStats(&user),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Update(r.Context(), &user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	stats := calculateUserStats(user)

	// Use a wrapper to include stats in the response
	response := struct {
		models.User
		Stats models.UserStats `json:"stats"`
	}{
		User:  *user,
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) GetHandHistory(w http.ResponseWriter, r *http.Request) {
	history, err := h.userRepo.GetPokerHistory(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch poker history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *AdminHandler) GetSlotHistory(w http.ResponseWriter, r *http.Request) {
	history, err := h.userRepo.GetSlotHistory(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch slot history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		Duration string `json:"duration"` // "1h", "24h", "permanent", etc.
		IsUnban  bool   `json:"is_unban"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, _ := primitive.ObjectIDFromHex(req.ID)
	user, err := h.userRepo.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if req.IsUnban {
		user.Status = "active"
		user.BannedUntil = 0
	} else if req.Duration == "permanent" {
		user.Status = "banned"
		user.BannedUntil = 0
	} else {
		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}
		user.Status = "banned" // Changed from "active" to "banned" for ban action
		user.BannedUntil = primitive.NewDateTimeFromTime(time.Now().Add(duration))
	}

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) AdjustBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Amount int64  `json:"amount"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, _ := primitive.ObjectIDFromHex(req.ID)
	if err := h.userRepo.UpdateBalance(r.Context(), id, req.Amount, req.Source); err != nil {
		http.Error(w, "Failed to adjust balance", http.StatusInternalServerError)
		return
	}

	h.hub.UpdateBalance(req.ID, req.Amount, req.Source)
	w.WriteHeader(http.StatusOK)
}
