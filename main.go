package main

import (
	"fmt"
	"log"
	"net/http"

	"bookstore/handlers"

	"github.com/gorilla/mux"
)

func main() {
	handlers.InitSampleData()

	bookHandler := handlers.NewBookHandler()
	authorHandler := handlers.NewAuthorHandler()
	categoryHandler := handlers.NewCategoryHandler()

	r := mux.NewRouter()

	r.HandleFunc("/books", bookHandler.ListBooks).Methods("GET")
	r.HandleFunc("/books", bookHandler.CreateBook).Methods("POST")
	r.HandleFunc("/books/{id:[0-9]+}", bookHandler.GetBook).Methods("GET")
	r.HandleFunc("/books/{id:[0-9]+}", bookHandler.UpdateBook).Methods("PUT")
	r.HandleFunc("/books/{id:[0-9]+}", bookHandler.DeleteBook).Methods("DELETE")

	r.HandleFunc("/authors", authorHandler.ListAuthors).Methods("GET")
	r.HandleFunc("/authors", authorHandler.CreateAuthor).Methods("POST")

	r.HandleFunc("/categories", categoryHandler.ListCategories).Methods("GET")
	r.HandleFunc("/categories", categoryHandler.CreateCategory).Methods("POST")

	port := ":8080"
	fmt.Printf("Server is running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
