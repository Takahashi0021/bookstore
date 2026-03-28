package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type Book struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	AuthorID   int     `json:"author_id"`
	CategoryID int     `json:"category_id"`
	Price      float64 `json:"price"`
}

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	Books      = make(map[int]Book)
	Authors    = make(map[int]Author)
	Categories = make(map[int]Category)
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

	filteredBooks := make([]Book, 0)
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

	book := Book{
		ID:         BookID,
		Title:      req.Title,
		AuthorID:   req.AuthorID,
		CategoryID: req.CategoryID,
		Price:      req.Price,
	}
	Books[BookID] = book
	BookID++

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

	book := Book{
		ID:         id,
		Title:      req.Title,
		AuthorID:   req.AuthorID,
		CategoryID: req.CategoryID,
		Price:      req.Price,
	}
	Books[id] = book

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
	w.WriteHeader(http.StatusNoContent)
}

func InitSampleData() {
	Authors[1] = Author{ID: 1, Name: "J.K. Rowling"}
	Authors[2] = Author{ID: 2, Name: "George R.R. Martin"}
	Authors[3] = Author{ID: 3, Name: "J.R.R. Tolkien"}
	Authors[4] = Author{ID: 4, Name: "Stephen King"}
	AuthorID = 5

	Categories[1] = Category{ID: 1, Name: "Fiction"}
	Categories[2] = Category{ID: 2, Name: "Fantasy"}
	Categories[3] = Category{ID: 3, Name: "Science Fiction"}
	Categories[4] = Category{ID: 4, Name: "Mystery"}
	Categories[5] = Category{ID: 5, Name: "Horror"}
	CategoryID = 6

	Books[1] = Book{
		ID:         1,
		Title:      "Harry Potter and the Philosopher's Stone",
		AuthorID:   1,
		CategoryID: 2,
		Price:      19.99,
	}
	Books[2] = Book{
		ID:         2,
		Title:      "A Game of Thrones",
		AuthorID:   2,
		CategoryID: 2,
		Price:      24.99,
	}
	Books[3] = Book{
		ID:         3,
		Title:      "The Hobbit",
		AuthorID:   3,
		CategoryID: 2,
		Price:      14.99,
	}
	Books[4] = Book{
		ID:         4,
		Title:      "The Shining",
		AuthorID:   4,
		CategoryID: 5,
		Price:      18.99,
	}
	BookID = 5
}
