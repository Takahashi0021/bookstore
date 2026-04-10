package main

import (
	"fmt"
	"log"
	"net/http"

	"bookstore/handlers"
	"bookstore/middleware"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	authHandler := handlers.NewAuthHandler()
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	bookHandler := handlers.NewBookHandler()
	authorHandler := handlers.NewAuthorHandler()
	categoryHandler := handlers.NewCategoryHandler()

	r.HandleFunc("/books", middleware.AuthMiddleware(bookHandler.ListBooks)).Methods("GET")
	r.HandleFunc("/books", middleware.AuthMiddleware(bookHandler.CreateBook)).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}", middleware.AuthMiddleware(bookHandler.GetBook)).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}", middleware.AuthMiddleware(bookHandler.UpdateBook)).Methods("PUT")
	r.HandleFunc("/books/{id:[0-9]+}", middleware.AuthMiddleware(bookHandler.DeleteBook)).Methods("DELETE")

	r.HandleFunc("/authors", middleware.AuthMiddleware(authorHandler.ListAuthors)).Methods("GET")
	r.HandleFunc("/authors", middleware.AuthMiddleware(authorHandler.CreateAuthor)).Methods("POST")

	r.HandleFunc("/categories", middleware.AuthMiddleware(categoryHandler.ListCategories)).Methods("GET")
	r.HandleFunc("/categories", middleware.AuthMiddleware(categoryHandler.CreateCategory)).Methods("POST")

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
