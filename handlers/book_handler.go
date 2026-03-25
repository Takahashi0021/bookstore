package handlers

import (
	"bookstore/models"
	"bookstore/storage"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type BookHandler struct {
	storage *storage.Storage
}

func NewBookHandler(storage *storage.Storage) *BookHandler {
	return &BookHandler{storage: storage}
}

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

	return errors
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books := h.storage.GetBooks()

	category := r.URL.Query().Get("category")
	author := r.URL.Query().Get("author")
	minPrice := r.URL.Query().Get("min_price")
	maxPrice := r.URL.Query().Get("max_price")

	filteredBooks := make([]models.Book, 0)
	for _, book := range books {
		if category != "" {
			cat, exists := h.storage.GetCategoryByID(book.CategoryID)
			if !exists || cat.Name != category {
				continue
			}
		}

		if author != "" {
			auth, exists := h.storage.GetAuthorByID(book.AuthorID)
			if !exists || auth.Name != author {
				continue
			}
		}

		if minPrice != "" {
			min, _ := strconv.ParseFloat(minPrice, 64)
			if book.Price < min {
				continue
			}
		}

		if maxPrice != "" {
			max, _ := strconv.ParseFloat(maxPrice, 64)
			if book.Price > max {
				continue
			}
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
		models.Book
		AuthorName   string `json:"author_name"`
		CategoryName string `json:"category_name"`
	}

	response := make([]BookResponse, 0)
	for _, book := range paginatedBooks {
		response = append(response, BookResponse{
			Book:         book,
			AuthorName:   h.storage.GetAuthorName(book.AuthorID),
			CategoryName: h.storage.GetCategoryName(book.CategoryID),
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

	if _, exists := h.storage.GetAuthorByID(req.AuthorID); !exists {
		http.Error(w, "Author not found", http.StatusBadRequest)
		return
	}

	if _, exists := h.storage.GetCategoryByID(req.CategoryID); !exists {
		http.Error(w, "Category not found", http.StatusBadRequest)
		return
	}

	book := models.Book{
		Title:      req.Title,
		AuthorID:   req.AuthorID,
		CategoryID: req.CategoryID,
		Price:      req.Price,
	}

	createdBook := h.storage.CreateBook(book)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBook)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	book, exists := h.storage.GetBookByID(id)
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	type BookResponse struct {
		models.Book
		AuthorName   string `json:"author_name"`
		CategoryName string `json:"category_name"`
	}

	response := BookResponse{
		Book:         book,
		AuthorName:   h.storage.GetAuthorName(book.AuthorID),
		CategoryName: h.storage.GetCategoryName(book.CategoryID),
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

	if _, exists := h.storage.GetAuthorByID(req.AuthorID); !exists {
		http.Error(w, "Author not found", http.StatusBadRequest)
		return
	}

	if _, exists := h.storage.GetCategoryByID(req.CategoryID); !exists {
		http.Error(w, "Category not found", http.StatusBadRequest)
		return
	}

	book := models.Book{
		Title:      req.Title,
		AuthorID:   req.AuthorID,
		CategoryID: req.CategoryID,
		Price:      req.Price,
	}

	updatedBook, exists := h.storage.UpdateBook(id, book)
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedBook)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	if exists := h.storage.DeleteBook(id); !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
