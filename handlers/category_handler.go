package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"bookstore/models"
	"bookstore/utils"
)

type CategoryRequest struct {
	Name string `json:"name"`
}

func ValidateCategory(category CategoryRequest) map[string]string {
	errors := make(map[string]string)

	name := strings.TrimSpace(category.Name)
	if name == "" {
		errors["name"] = "Name is required"
	} else if len(name) < 2 {
		errors["name"] = "Name must be at least 2 characters long"
	} else if len(name) > 50 {
		errors["name"] = "Name cannot exceed 50 characters"
	}

	return errors
}

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	categoryList := make([]models.Category, 0, len(Categories))
	for _, category := range Categories {
		categoryList = append(categoryList, category)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  categoryList,
		"total": len(categoryList),
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

	mu.Lock()
	defer mu.Unlock()

	for _, existingCategory := range Categories {
		if strings.EqualFold(existingCategory.Name, strings.TrimSpace(req.Name)) {
			http.Error(w, "Category already exists", http.StatusConflict)
			return
		}
	}

	category := models.Category{
		ID:   CategoryID,
		Name: strings.TrimSpace(req.Name),
	}
	Categories[CategoryID] = category
	CategoryID++

	go utils.SaveCategories(Categories)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}
