package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"loadbalancer/database"
	"loadbalancer/server"
	"log"
	"os"
	"strconv"
	"sync"
)

var conn redis.Conn

func main() {
	portStr := os.Getenv("LOADBALANCER_PORT")
	pool := database.GetDatabase()
	conn = pool.Get()
	defer conn.Close()
	if portStr == "" {
		panic("Port wasn't found\n")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err)
		panic(fmt.Sprintf("error converting port from string to int : %s", err.Error()))
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		server.StartUDP("0.0.0.0", port)
	}()
	go func() {
		defer wg.Done()
		startTcpServer(port)
	}()
	wg.Wait()
}

func startTcpServer(port int) {
	err := server.StartTCP(port)
	if err != nil {
		log.Printf("Error while running tcp server %v", err)
	}
}