package main

import (
	"log"
	"net/http"
	"os"
	"supervisor/routers"
)

func main() {
	http.Handle("/", routers.InitRoutes())

	port := os.Getenv("SUPERVISOR_PORT")
	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
