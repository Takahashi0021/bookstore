package handlers

import (
	"bookstore/models"
	"bookstore/storage"
	"encoding/json"
	"net/http"
	"strings"
)

type AuthorHandler struct {
	storage *storage.Storage
}

func NewAuthorHandler(storage *storage.Storage) *AuthorHandler {
	return &AuthorHandler{storage: storage}
}

type AuthorRequest struct {
	Name string `json:"name"`
}

func ValidateAuthor(author AuthorRequest) map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(author.Name) == "" {
		errors["name"] = "Name is required"
	}

	return errors
}

func (h *AuthorHandler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	authors := h.storage.GetAuthors()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  authors,
		"total": len(authors),
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

	author := models.Author{
		Name: strings.TrimSpace(req.Name),
	}

	createdAuthor := h.storage.CreateAuthor(author)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdAuthor)
}
