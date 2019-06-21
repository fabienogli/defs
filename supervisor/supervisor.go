package main

import (
	"log"
	"net/http"
	"supervisor/utils"
	"supervisor/routers"
)

func main() {
	http.Handle("/", routers.InitRoutes())

	port := utils.GetRequiredEnv("SUPERVISOR_PORT")
	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
