package main

import (
	"log"
	"net/http"
	"supervisor/routers"
)

func main() {
	http.Handle("/", routers.InitRoutes())
	log.Println("Listening on 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}