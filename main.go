package main

import (
	"fmt"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello World</h1>")
}

func main() {
	fmt.Println("Server is listening...\n Server used port:3000")
	http.HandlerFunc("/", indexHandler)
	http.ListenAndServe(":3000", nil)
}
