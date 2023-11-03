package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./html/")))

	http.HandleFunc("/hello", sayHello)

	fmt.Println("Starting server at port 8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}

func sayHello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, Welcome to the internal Gig")
}
