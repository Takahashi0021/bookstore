package main

import (
	"bookstore/handlers"
	"bookstore/models"
	"bookstore/storage"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	store := storage.NewStorage()

	bookHandler := handlers.NewBookHandler(store)
	authorHandler := handlers.NewAuthorHandler(store)
	categoryHandler := handlers.NewCategoryHandler(store)

	r := mux.NewRouter()

	r.HandleFunc("/books", bookHandler.ListBooks).Methods("GET")
	r.HandleFunc("/books", bookHandler.CreateBook).Methods("POST")
	r.HandleFunc("/books/{id}", bookHandler.GetBook).Methods("GET")
	r.HandleFunc("/books/{id}", bookHandler.UpdateBook).Methods("PUT")
	r.HandleFunc("/books/{id}", bookHandler.DeleteBook).Methods("DELETE")

	r.HandleFunc("/authors", authorHandler.ListAuthors).Methods("GET")
	r.HandleFunc("/authors", authorHandler.CreateAuthor).Methods("POST")

	r.HandleFunc("/categories", categoryHandler.ListCategories).Methods("GET")
	r.HandleFunc("/categories", categoryHandler.CreateCategory).Methods("POST")

	addSampleData(store)

	port := ":8080"
	fmt.Printf("Server is running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

func addSampleData(store *storage.Storage) {
	store.CreateAuthor(models.Author{Name: "J.K. Rowling"})
	store.CreateAuthor(models.Author{Name: "George R.R. Martin"})
	store.CreateAuthor(models.Author{Name: "J.R.R. Tolkien"})

	store.CreateCategory(models.Category{Name: "Fiction"})
	store.CreateCategory(models.Category{Name: "Fantasy"})
	store.CreateCategory(models.Category{Name: "Science Fiction"})
	store.CreateCategory(models.Category{Name: "Mystery"})

	store.CreateBook(models.Book{
		Title:      "Harry Potter and the Philosopher's Stone",
		AuthorID:   1,
		CategoryID: 2,
		Price:      19.99,
	})
	store.CreateBook(models.Book{
		Title:      "A Game of Thrones",
		AuthorID:   2,
		CategoryID: 2,
		Price:      24.99,
	})
	store.CreateBook(models.Book{
		Title:      "The Hobbit",
		AuthorID:   3,
		CategoryID: 2,
		Price:      14.99,
	})
}
