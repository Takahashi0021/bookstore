package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"bookstore/models"
	"bookstore/utils"

	"github.com/gorilla/mux"
)

var (
	Books      = make(map[int]models.Book)
	Authors    = make(map[int]models.Author)
	Categories = make(map[int]models.Category)
	BookID     = 1
	AuthorID   = 1
	CategoryID = 1
	mu         sync.RWMutex
)

type BookRequest struct {
	Title      string  `json:"title"`
	AuthorID   int     `json:"author_id"`
	CategoryID int     `json:"category_id"`
	Price      float64 `json:"price"`
}

func ValidateBook(book BookRequest) map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(book.Title) == "" {
		errors["title"] = "Title is required"
	}

	if book.AuthorID <= 0 {
		errors["author_id"] = "Author ID must be a positive integer"
	}

	if book.CategoryID <= 0 {
		errors["category_id"] = "Category ID must be a positive integer"
	}

	if book.Price < 0 {
		errors["price"] = "Price cannot be negative"
	}

	if book.Price > 10000 {
		errors["price"] = "Price cannot exceed 10000"
	}

	return errors
}

type BookHandler struct{}

func NewBookHandler() *BookHandler {
	return &BookHandler{}
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	categoryName := r.URL.Query().Get("category")
	authorName := r.URL.Query().Get("author")
	minPriceStr := r.URL.Query().Get("min_price")
	maxPriceStr := r.URL.Query().Get("max_price")

	var minPrice, maxPrice float64
	var err error

	if minPriceStr != "" {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			http.Error(w, "Invalid min_price", http.StatusBadRequest)
			return
		}
	}

	if maxPriceStr != "" {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			http.Error(w, "Invalid max_price", http.StatusBadRequest)
			return
		}
	}

	filteredBooks := make([]models.Book, 0)
	for _, book := range Books {
		if categoryName != "" {
			category, exists := Categories[book.CategoryID]
			if !exists || !strings.EqualFold(category.Name, categoryName) {
				continue
			}
		}

		if authorName != "" {
			author, exists := Authors[book.AuthorID]
			if !exists || !strings.EqualFold(author.Name, authorName) {
				continue
			}
		}

		if minPriceStr != "" && book.Price < minPrice {
			continue
		}
		if maxPriceStr != "" && book.Price > maxPrice {
			continue
		}

		filteredBooks = append(filteredBooks, book)
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

	if start > len(filteredBooks) {
		start = len(filteredBooks)
	}
	if end > len(filteredBooks) {
		end = len(filteredBooks)
	}

	paginatedBooks := filteredBooks[start:end]

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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":        response,
		"page":        page,
		"page_size":   pageSize,
		"total":       len(filteredBooks),
		"total_pages": (len(filteredBooks) + pageSize - 1) / pageSize,
	})
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req BookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if errors := ValidateBook(req); len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := Authors[req.AuthorID]; !exists {
		http.Error(w, "Author not found", http.StatusBadRequest)
		return
	}

	if _, exists := Categories[req.CategoryID]; !exists {
		http.Error(w, "Category not found", http.StatusBadRequest)
		return
	}

	book := models.Book{
		ID:         BookID,
		Title:      req.Title,
		AuthorID:   req.AuthorID,
		CategoryID: req.CategoryID,
		Price:      req.Price,
	}
	Books[BookID] = book
	BookID++

	go utils.SaveBooks(Books)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	book, exists := Books[id]
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	authorName := "Unknown"
	if author, exists := Authors[book.AuthorID]; exists {
		authorName = author.Name
	}

	categoryName := "Unknown"
	if category, exists := Categories[book.CategoryID]; exists {
		categoryName = category.Name
	}

	type BookResponse struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		AuthorID     int     `json:"author_id"`
		AuthorName   string  `json:"author_name"`
		CategoryID   int     `json:"category_id"`
		CategoryName string  `json:"category_name"`
		Price        float64 `json:"price"`
	}

	response := BookResponse{
		ID:           book.ID,
		Title:        book.Title,
		AuthorID:     book.AuthorID,
		AuthorName:   authorName,
		CategoryID:   book.CategoryID,
		CategoryName: categoryName,
		Price:        book.Price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	var req BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if errors := ValidateBook(req); len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := Books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	if _, exists := Authors[req.AuthorID]; !exists {
		http.Error(w, "Author not found", http.StatusBadRequest)
		return
	}

	if _, exists := Categories[req.CategoryID]; !exists {
		http.Error(w, "Category not found", http.StatusBadRequest)
		return
	}

	book := models.Book{
		ID:         id,
		Title:      req.Title,
		AuthorID:   req.AuthorID,
		CategoryID: req.CategoryID,
		Price:      req.Price,
	}
	Books[id] = book

	go utils.SaveBooks(Books)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := Books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(Books, id)

	go utils.SaveBooks(Books)

	w.WriteHeader(http.StatusNoContent)
}

func SaveAllData() {
	utils.SaveBooks(Books)
	utils.SaveAuthors(Authors)
	utils.SaveCategories(Categories)
}

func LoadAllData() {
	if loadedBooks, err := utils.LoadBooks(); err == nil && len(loadedBooks) > 0 {
		Books = loadedBooks
		maxID := 0
		for id := range Books {
			if id > maxID {
				maxID = id
			}
		}
		BookID = maxID + 1
	}

	if loadedAuthors, err := utils.LoadAuthors(); err == nil && len(loadedAuthors) > 0 {
		Authors = loadedAuthors
		maxID := 0
		for id := range Authors {
			if id > maxID {
				maxID = id
			}
		}
		AuthorID = maxID + 1
	}

	if loadedCategories, err := utils.LoadCategories(); err == nil && len(loadedCategories) > 0 {
		Categories = loadedCategories
		maxID := 0
		for id := range Categories {
			if id > maxID {
				maxID = id
			}
		}
		CategoryID = maxID + 1
	}
}
