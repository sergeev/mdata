package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

// Account sql structure
// CREATE TABLE users(
//		id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
//  	username VARCHAR(50),
// password VARCHAR(120)
// );

var db *sql.DB
var err error

func main() {
	// default load template
	templates = template.Must(template.ParseGlob("app/views/layouts/*.html"))
	// account load template
	//templates = aTemplate.Must(template.ParseGlob("app/views/layouts/account/*.html"))


	// Load database
	// db, err = sql.Open("mysql", "myUsername:myPassword@/myDatabase")
	db, err = sql.Open("mysql", "root:_@/golang_db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	//handler := http.NewServeMux()

	// C R U D
	//handler.HandleFunc("/", Logger(indexHandler))
	//handler.HandleFunc("/hello/", Logger(BasicAuth(helloHandler)))

	//handler.HandleFunc("/book/", Logger(bookHandler))

	//handler.HandleFunc("/books/", Logger(booksHandler))

	r := mux.NewRouter()
	// main deep router
	r.HandleFunc("/", Logger(indexHandler))

	// administrative router
	r.HandleFunc("/manage", Logger(BasicAuth(indexManage)))

	// register and login route
	r.HandleFunc("/signup", Logger(signupPage))
	r.HandleFunc("/login", Logger(loginPage))

	// misc router
	r.HandleFunc("/hello/", Logger(BasicAuth(helloHandler)))
	r.HandleFunc("/book/", Logger(bookHandler))
	r.HandleFunc("/books/", Logger(booksHandler))

	http.Handle("/", r)

	// server configs
	s := http.Server{
		Addr:           "0.0.0.0:8000",
		//Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// Test log server witch fatal error
	log.Fatal(s.ListenAndServe())
}

type Resp struct {
	// resp structure
	Message interface{}
	Error   string
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	name := strings.Replace(r.URL.Path, "/hello/", "", 1)

	resp := Resp{
		Message: fmt.Sprintf("hello %s. Glad to see you again", name),
	}

	respJson, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)

	w.Write(respJson)
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("Server [net/http] method [%s] connection from [%v]", r.Method, r.RemoteAddr)

		next.ServeHTTP(w, r)
	}
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGetBook(w, r)
	} else if r.Method == http.MethodPost {
		handleAddBook(w, r)
	} else if r.Method == http.MethodDelete {
		handleDeleteBook(w, r)
	} else if r.Method == http.MethodPut {
		handleUpdateBook(w, r)
	}
}

func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)

			return
		}

		hashed, _ := base64.StdEncoding.DecodeString(auth[1])

		pair := strings.SplitN(string(hashed), ":", 2)

		log.Printf("pair %+v", pair)

		if len(pair) != 2 || !aAuth(pair[0], pair[1]) {
			http.Error(w, "Authorization failed!", http.StatusUnauthorized)

			return
		}
		next.ServeHTTP(w, r)
	}
}

func aAuth(username, password string) bool {
	if username == "test" && password == "test" {
		return true
	}
	return false
}

func handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)
	decoder := json.NewDecoder(r.Body)

	var book Book

	var resp Resp

	err := decoder.Decode(&book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		return
	}

	book.Id = id

	err = bookStore.UpdateBook(book)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = fmt.Sprintf("")
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		return
	}
	resp.Message = book

	respJson, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	var resp Resp

	err := bookStore.DeleteBook(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = fmt.Sprintf("")
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		return
	}
	booksHandler(w, r)
}

func handleAddBook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var book Book

	var resp Resp

	err := decoder.Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		return
	}

	err = bookStore.AddBooks(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		return
	}
	booksHandler(w, r)
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handleGetBook(w, r)
	}
	w.WriteHeader(http.StatusOK)
	resp := Resp{
		Message: bookStore.GetBooks(),
	}
	booksJson, _ := json.Marshal(resp)
	w.Write(booksJson)
}

func handleGetBook(w http.ResponseWriter, r *http.Request) {
	id := strings.Replace(r.URL.Path, "/book/", "", 1)

	var resp Resp

	book := bookStore.FindBookById(id)
	if book == nil {
		w.WriteHeader(http.StatusNotFound)
		resp.Error = fmt.Sprintf("")
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		return
	}
	resp.Message = book

	respJson, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

type Book struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Name   string `json:"name"`
}

type BookStore struct {
	books []Book
}

var bookStore = BookStore{
	books: make([]Book, 0),
}

func (s BookStore) FindBookById(id string) *Book {
	for _, book := range s.books {
		if book.Id == id {
			return &book
		}
	}
	return nil
}

func (s BookStore) GetBooks() []Book {
	return s.books
}

func (s *BookStore) AddBooks(book Book) error {
	for _, bk := range s.books {
		if bk.Id == book.Id {
			return errors.New(fmt.Sprintf("Book witch id %s not found", book.Id))
		}
	}
	s.books = append(s.books, book)
	return nil
}

func (s *BookStore) UpdateBook(book Book) error {
	for i, bk := range s.books {
		if bk.Id == book.Id {
			s.books[i] = book
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Book with id %s not found", book.Id))
}

func (s *BookStore) DeleteBook(id string) error {
	for i, bk := range s.books {
		if bk.Id == id {
			s.books = append(s.books[:i], s.books[i+1:]...)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Book with id %s not found", id))
}

func signupPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		//http.ServeFile(res, req, "signup.html")
		http.ServeFile(res, req, "/signup.html")
		//templates.ExecuteTemplate(res, "account/signup.html", nil)
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var user string

	err := db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)

	switch {
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		_, err = db.Exec("INSERT INTO users(username, password) VALUES(?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}

		res.Write([]byte("User created!"))
		return
	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "app/views/layouts/account/login.html")

		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var databaseUsername string
	var databasePassword string

	err := db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword)

	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}

	res.Write([]byte("Hello" + databaseUsername))

}

// Template parse
var templates *template.Template

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func indexManage(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "manage.html", nil)
}
