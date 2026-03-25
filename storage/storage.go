package storage

import (
	"bookstore/models"
	"fmt"
	"sync"
)

type Storage struct {
	mu         sync.RWMutex
	books      map[int]models.Book
	authors    map[int]models.Author
	categories map[int]models.Category
	bookID     int
	authorID   int
	categoryID int
}

func NewStorage() *Storage {
	return &Storage{
		books:      make(map[int]models.Book),
		authors:    make(map[int]models.Author),
		categories: make(map[int]models.Category),
		bookID:     1,
		authorID:   1,
		categoryID: 1,
	}
}

func (s *Storage) CreateBook(book models.Book) models.Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	book.ID = s.bookID
	s.books[s.bookID] = book
	s.bookID++
	return book
}

func (s *Storage) GetBooks() []models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	books := make([]models.Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	return books
}

func (s *Storage) GetBookByID(id int) (models.Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, exists := s.books[id]
	return book, exists
}

func (s *Storage) UpdateBook(id int, book models.Book) (models.Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.books[id]; !exists {
		return models.Book{}, false
	}

	book.ID = id
	s.books[id] = book
	return book, true
}

func (s *Storage) DeleteBook(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.books[id]; !exists {
		return false
	}

	delete(s.books, id)
	return true
}

func (s *Storage) CreateAuthor(author models.Author) models.Author {
	s.mu.Lock()
	defer s.mu.Unlock()

	author.ID = s.authorID
	s.authors[s.authorID] = author
	s.authorID++
	return author
}

func (s *Storage) GetAuthors() []models.Author {
	s.mu.RLock()
	defer s.mu.RUnlock()

	authors := make([]models.Author, 0, len(s.authors))
	for _, author := range s.authors {
		authors = append(authors, author)
	}
	return authors
}

func (s *Storage) GetAuthorByID(id int) (models.Author, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	author, exists := s.authors[id]
	return author, exists
}

func (s *Storage) CreateCategory(category models.Category) models.Category {
	s.mu.Lock()
	defer s.mu.Unlock()

	category.ID = s.categoryID
	s.categories[s.categoryID] = category
	s.categoryID++
	return category
}

func (s *Storage) GetCategories() []models.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categories := make([]models.Category, 0, len(s.categories))
	for _, category := range s.categories {
		categories = append(categories, category)
	}
	return categories
}

func (s *Storage) GetCategoryByID(id int) (models.Category, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	category, exists := s.categories[id]
	return category, exists
}

func (s *Storage) GetAuthorName(id int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if author, exists := s.authors[id]; exists {
		return author.Name
	}
	return fmt.Sprintf("Author ID: %d", id)
}

func (s *Storage) GetCategoryName(id int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if category, exists := s.categories[id]; exists {
		return category.Name
	}
	return fmt.Sprintf("Category ID: %d", id)
}
