package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AuthorRequest struct {
	Name string `json:"name"`
}

func ValidateAuthor(author AuthorRequest) map[string]string {
	errors := make(map[string]string)

	name := strings.TrimSpace(author.Name)
	if name == "" {
		errors["name"] = "Name is required"
	} else if len(name) < 2 {
		errors["name"] = "Name must be at least 2 characters long"
	} else if len(name) > 100 {
		errors["name"] = "Name cannot exceed 100 characters"
	}

	return errors
}

type AuthorHandler struct{}

func NewAuthorHandler() *AuthorHandler {
	return &AuthorHandler{}
}

func (h *AuthorHandler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	authorList := make([]Author, 0, len(Authors))
	for _, author := range Authors {
		authorList = append(authorList, author)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  authorList,
		"total": len(authorList),
	})
}

func (h *AuthorHandler) CreateAuthor(w http.ResponseWriter, r *http.Request) {
	var req AuthorRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if errors := ValidateAuthor(req); len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for _, existingAuthor := range Authors {
		if strings.EqualFold(existingAuthor.Name, strings.TrimSpace(req.Name)) {
			http.Error(w, "Author with this name already exists", http.StatusConflict)
			return
		}
	}

	author := Author{
		ID:   AuthorID,
		Name: strings.TrimSpace(req.Name),
	}
	Authors[AuthorID] = author
	AuthorID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(author)
}
