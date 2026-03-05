package handlers

import (
	"encoding/json"
	"fmt"
	"kashino-backend/internal/mail"
	"kashino-backend/internal/models"
	"kashino-backend/internal/repository"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type UserHandler struct {
	repo    *repository.UserRepository
	mailSvc *mail.MailService
}

func NewUserHandler(repo *repository.UserRepository, mailSvc *mail.MailService) *UserHandler {
	if len(jwtKey) == 0 {
		jwtKey = []byte("default_secret") // For development only
	}
	return &UserHandler{repo: repo, mailSvc: mailSvc}
}

func (h *UserHandler) RequestSignupOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if username or email already exists in a verified account
	existingUser, _ := h.repo.FindByUsername(r.Context(), req.Username)
	if existingUser != nil && existingUser.IsVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username already taken"})
		return
	}

	existingEmail, _ := h.repo.FindByEmail(r.Context(), req.Email)
	if existingEmail != nil && existingEmail.IsVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email already registered"})
		return
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiry := time.Now().Add(10 * time.Minute)

	// We can use the existing UpdateOTP, but since the user doesn't exist yet,
	// we'll either need to create a "pending" user or use a dedicated OTP collection.
	// For simplicity, let's look for a pending user or create one.
	pendingUser := existingEmail
	if pendingUser == nil {
		pendingUser = &models.User{
			Username:   req.Username,
			Email:      req.Email,
			IsVerified: false,
		}
		if err := h.repo.Create(r.Context(), pendingUser); err != nil {
			http.Error(w, "Failed to initiate signup", http.StatusInternalServerError)
			return
		}
	} else {
		// Update username in case they changed it
		pendingUser.Username = req.Username
		h.repo.Update(r.Context(), pendingUser)
	}

	if err := h.repo.UpdateOTP(r.Context(), req.Email, otp, expiry); err != nil {
		http.Error(w, "Failed to generate OTP", http.StatusInternalServerError)
		return
	}

	if err := h.mailSvc.SendOTP(req.Email, otp); err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent successfully"})
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		OTP      string `json:"otp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify OTP
	user, err := h.repo.FindByEmail(r.Context(), req.Email)
	if err != nil || user.OTP != req.OTP || time.Now().After(user.OTPExpiry.Time()) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired OTP"})
		return
	}

	// Double check username availability (if they changed it after last OTP)
	existing, _ := h.repo.FindByUsername(r.Context(), req.Username)
	if existing != nil && existing.ID != user.ID && existing.IsVerified {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// Update existing pending user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user.Username = req.Username
	user.Password = string(hashedPassword)
	user.IsVerified = true
	user.OTP = ""
	user.OTPExpiry = 0

	if err := h.repo.Update(r.Context(), user); err != nil {
		http.Error(w, "Failed to finalize signup", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.repo.FindByUsername(r.Context(), credentials.Username)
	if err != nil {
		fmt.Printf("SignIn failure: user '%s' not found\n", credentials.Username)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if user.BannedUntil > 0 {
		bannedUntil := user.BannedUntil.Time()
		if time.Now().Before(bannedUntil) {
			fmt.Printf("SignIn failure: user '%s' is banned until %v\n", user.Username, bannedUntil)
			http.Error(w, fmt.Sprintf("Account is banned until %v", bannedUntil.Format("2006-01-02 15:04:05")), http.StatusForbidden)
			return
		}
	}

	if user.Status == "banned" {
		fmt.Printf("SignIn failure: user '%s' is permanently banned\n", user.Username)
		http.Error(w, "Account is permanently banned", http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		fmt.Printf("SignIn failure: password mismatch for user '%s'\n", credentials.Username)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if user.Status == "banned" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Your account has been banned. Please contact support."})
		return
	}

	if !user.IsVerified {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Account not verified. Please check your email for OTP."})
		return
	}

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"sub":  user.ID.Hex(),
		"role": user.Role,
		"exp":  expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Recalculate level and tier from exp
	level := models.CalculateLevel(user.Exp)
	tier := models.GetTierFromLevel(level)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":           tokenString,
		"id":              user.ID.Hex(),
		"username":        user.Username,
		"balance":         user.Balance,
		"tier":            tier,
		"exp":             user.Exp,
		"level":           level,
		"profile_picture": user.ProfilePicture,
		"role":            user.Role,
	})
}

func (h *UserHandler) GetSlotHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history, err := h.repo.GetSlotHistory(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch slot history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found or error fetching data", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  idStr,
		"balance":  user.Balance,
		"history":  user.BalanceHistory,
		"username": user.Username,
		"tier":     user.Tier,
	})
}

func (h *UserHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiry := time.Now().Add(10 * time.Minute)

	if err := h.repo.UpdateOTP(r.Context(), req.Email, otp, expiry); err != nil {
		http.Error(w, "Failed to update OTP", http.StatusInternalServerError)
		return
	}

	if err := h.mailSvc.SendOTP(req.Email, otp); err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent successfully"})
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token := primitive.NewObjectID().Hex()
	expiry := time.Now().Add(1 * time.Hour)

	if err := h.repo.UpdateResetToken(r.Context(), req.Email, token, expiry); err != nil {
		http.Error(w, "Failed to update reset token", http.StatusInternalServerError)
		return
	}

	if err := h.mailSvc.SendResetPassword(req.Email, token); err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Reset token sent successfully"})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.repo.FindByResetToken(r.Context(), req.Token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	if time.Now().After(user.ResetTokenExpiry.Time()) {
		http.Error(w, "Token expired", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	user.ResetToken = ""
	user.ResetTokenExpiry = 0

	if err := h.repo.Update(r.Context(), user); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}
