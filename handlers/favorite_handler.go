package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"bookstore/middleware"
	"bookstore/models"
	"bookstore/utils"

	"github.com/gorilla/mux"
)

type FavoriteHandler struct{}

func NewFavoriteHandler() *FavoriteHandler {
	return &FavoriteHandler{}
}

var favorites = make(map[int]map[int]models.Favorite)

func init() {
	favorites = make(map[int]map[int]models.Favorite)
}

func (h *FavoriteHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["bookId"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := Books[bookID]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	if favorites[user.UserID] == nil {
		favorites[user.UserID] = make(map[int]models.Favorite)
	}

	if _, exists := favorites[user.UserID][bookID]; exists {
		http.Error(w, "Book already in favorites", http.StatusConflict)
		return
	}

	favorite := models.Favorite{
		UserID:    user.UserID,
		BookID:    bookID,
		CreatedAt: time.Now(),
	}

	favorites[user.UserID][bookID] = favorite

	go utils.SaveFavorites(favorites)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Book added to favorites successfully",
		"favorite": favorite,
	})
}

func (h *FavoriteHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["bookId"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if favorites[user.UserID] == nil {
		http.Error(w, "Book not in favorites", http.StatusNotFound)
		return
	}

	if _, exists := favorites[user.UserID][bookID]; !exists {
		http.Error(w, "Book not in favorites", http.StatusNotFound)
		return
	}

	delete(favorites[user.UserID], bookID)

	go utils.SaveFavorites(favorites)

	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoriteHandler) GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	favoriteBooks := make([]models.Book, 0)

	if userFavorites, exists := favorites[user.UserID]; exists {
		for bookID := range userFavorites {
			if book, exists := Books[bookID]; exists {
				favoriteBooks = append(favoriteBooks, book)
			}
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(favoriteBooks) {
		start = len(favoriteBooks)
	}
	if end > len(favoriteBooks) {
		end = len(favoriteBooks)
	}

	paginatedBooks := favoriteBooks[start:end]

	type BookResponse struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		AuthorID     int     `json:"author_id"`
		AuthorName   string  `json:"author_name"`
		CategoryID   int     `json:"category_id"`
		CategoryName string  `json:"category_name"`
		Price        float64 `json:"price"`
	}

	response := make([]BookResponse, 0)
	for _, book := range paginatedBooks {
		authorName := "Unknown"
		if author, exists := Authors[book.AuthorID]; exists {
			authorName = author.Name
		}

		categoryName := "Unknown"
		if category, exists := Categories[book.CategoryID]; exists {
			categoryName = category.Name
		}

		response = append(response, BookResponse{
			ID:           book.ID,
			Title:        book.Title,
			AuthorID:     book.AuthorID,
			AuthorName:   authorName,
			CategoryID:   book.CategoryID,
			CategoryName: categoryName,
			Price:        book.Price,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":        response,
		"page":        page,
		"page_size":   pageSize,
		"total":       len(favoriteBooks),
		"total_pages": (len(favoriteBooks) + pageSize - 1) / pageSize,
	})
}
