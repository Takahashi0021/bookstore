package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"bookstore/models"
	"bookstore/utils"
)

var (
	Users  = make(map[int]models.User)
	UserID = 1
	muAuth sync.RWMutex
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errors := make(map[string]string)
	if strings.TrimSpace(req.Username) == "" {
		errors["username"] = "Username is required"
	}
	if len(req.Username) < 3 {
		errors["username"] = "Username must be at least 3 characters"
	}
	if strings.TrimSpace(req.Password) == "" {
		errors["password"] = "Password is required"
	}
	if len(req.Password) < 6 {
		errors["password"] = "Password must be at least 6 characters"
	}

	if len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"errors": errors})
		return
	}

	muAuth.Lock()
	defer muAuth.Unlock()

	for _, existingUser := range Users {
		if existingUser.Username == req.Username {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
	}

	user := models.User{
		ID:       UserID,
		Username: req.Username,
		Password: req.Password,
		Role:     "user",
	}

	Users[UserID] = user
	UserID++

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	user.Password = ""

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	muAuth.RLock()
	defer muAuth.RUnlock()

	var foundUser *models.User
	for _, user := range Users {
		if user.Username == req.Username {
			foundUser = &user
			break
		}
	}

	if foundUser == nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if foundUser.Password != req.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(foundUser.ID, foundUser.Username, foundUser.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	foundUser.Password = ""

	response := models.LoginResponse{
		Token: token,
		User:  *foundUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
