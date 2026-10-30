package library

type Book struct {
	Title string `json:"title"`
}

type Shelf struct {
	Books []Book `json:"books"`
}

func (shelf Shelf) Add(book Book) Shelf {
	return Shelf{Books: append(append([]Book(nil), shelf.Books...), book)}
}
