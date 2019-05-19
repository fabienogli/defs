package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from the swag!")
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}