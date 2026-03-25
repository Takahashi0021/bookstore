package handlers

import (
	"bookstore/models"
	"bookstore/storage"
	"encoding/json"
	"net/http"
	"strings"
)

type CategoryHandler struct {
	storage *storage.Storage
}

func NewCategoryHandler(storage *storage.Storage) *CategoryHandler {
	return &CategoryHandler{storage: storage}
}

type CategoryRequest struct {
	Name string `json:"name"`
}

func ValidateCategory(category CategoryRequest) map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(category.Name) == "" {
		errors["name"] = "Name is required"
	}

	return errors
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories := h.storage.GetCategories()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  categories,
		"total": len(categories),
	})
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if errors := ValidateCategory(req); len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	category := models.Category{
		Name: strings.TrimSpace(req.Name),
	}

	createdCategory := h.storage.CreateCategory(category)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCategory)
}
