package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bookstore/handlers"
	"bookstore/middleware"
	"bookstore/models"
	"bookstore/utils"

	"github.com/gorilla/mux"
)

func main() {
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal("Cannot create data directory:", err)
	}

	handlers.LoadAllData()

	if len(handlers.Books) == 0 {
		initSampleData()
	}

	if len(handlers.Users) == 0 {
		initTestUser()
	}

	r := mux.NewRouter()

	authHandler := handlers.NewAuthHandler()
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	bookHandler := handlers.NewBookHandler()
	authorHandler := handlers.NewAuthorHandler()
	categoryHandler := handlers.NewCategoryHandler()
	favoriteHandler := handlers.NewFavoriteHandler()

	r.HandleFunc("/books", middleware.AuthMiddleware(bookHandler.ListBooks)).Methods("GET")
	r.HandleFunc("/books", middleware.AuthMiddleware(bookHandler.CreateBook)).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}", middleware.AuthMiddleware(bookHandler.GetBook)).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}", middleware.AuthMiddleware(bookHandler.UpdateBook)).Methods("PUT")
	r.HandleFunc("/books/{id:[0-9]+}", middleware.AuthMiddleware(bookHandler.DeleteBook)).Methods("DELETE")

	r.HandleFunc("/books/favorites", middleware.AuthMiddleware(favoriteHandler.GetUserFavorites)).Methods("GET")
	r.HandleFunc("/books/{bookId:[0-9]+}/favorites", middleware.AuthMiddleware(favoriteHandler.AddFavorite)).Methods("PUT")
	r.HandleFunc("/books/{bookId:[0-9]+}/favorites", middleware.AuthMiddleware(favoriteHandler.RemoveFavorite)).Methods("DELETE")

	r.HandleFunc("/authors", middleware.AuthMiddleware(authorHandler.ListAuthors)).Methods("GET")
	r.HandleFunc("/authors", middleware.AuthMiddleware(authorHandler.CreateAuthor)).Methods("POST")

	r.HandleFunc("/categories", middleware.AuthMiddleware(categoryHandler.ListCategories)).Methods("GET")
	r.HandleFunc("/categories", middleware.AuthMiddleware(categoryHandler.CreateCategory)).Methods("POST")

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

func initSampleData() {
	handlers.Authors[1] = models.Author{ID: 1, Name: "J.K. Rowling"}
	handlers.Authors[2] = models.Author{ID: 2, Name: "George R.R. Martin"}
	handlers.Authors[3] = models.Author{ID: 3, Name: "J.R.R. Tolkien"}
	handlers.AuthorID = 4

	handlers.Categories[1] = models.Category{ID: 1, Name: "Fiction"}
	handlers.Categories[2] = models.Category{ID: 2, Name: "Fantasy"}
	handlers.Categories[3] = models.Category{ID: 3, Name: "Science Fiction"}
	handlers.Categories[4] = models.Category{ID: 4, Name: "Mystery"}
	handlers.CategoryID = 5

	handlers.Books[1] = models.Book{
		ID:         1,
		Title:      "Harry Potter and the Philosopher's Stone",
		AuthorID:   1,
		CategoryID: 2,
		Price:      19.99,
	}
	handlers.Books[2] = models.Book{
		ID:         2,
		Title:      "A Game of Thrones",
		AuthorID:   2,
		CategoryID: 2,
		Price:      24.99,
	}
	handlers.Books[3] = models.Book{
		ID:         3,
		Title:      "The Hobbit",
		AuthorID:   3,
		CategoryID: 2,
		Price:      14.99,
	}
	handlers.BookID = 4

	handlers.SaveAllData()
}

func initTestUser() {
	handlers.Users[1] = models.User{
		ID:       1,
		Username: "admin",
		Password: "admin123",
		Role:     "admin",
	}
	handlers.Users[2] = models.User{
		ID:       2,
		Username: "user",
		Password: "user123",
		Role:     "user",
	}
	handlers.UserID = 3

	utils.SaveUsers(handlers.Users)
}
