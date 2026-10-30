package main

import (
	"fmt"
	"os"

	"bookshelf/internal/library"
	"bookshelf/internal/storage"
)

func main() {
	store := storage.Open("books.json")
	shelf, err := store.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "bookshelf:", err)
		os.Exit(1)
	}
	if len(os.Args) == 3 && os.Args[1] == "add" {
		shelf = shelf.Add(library.Book{Title: os.Args[2]})
		if err := store.Save(shelf); err != nil {
			fmt.Fprintln(os.Stderr, "bookshelf:", err)
			os.Exit(1)
		}
	}
	for _, book := range shelf.Books {
		fmt.Println(book.Title)
	}
}
