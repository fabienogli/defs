package main

import (
	"net"
	"log"
	"loadbalancer/database"
	"errors"
	"os"
)

func main() {
	port := os.Getenv("port")
	if port == "" {
		panic("Port wasn't found\n")
	}
	port = ":" + port
	log.Printf("Listening on %s\n", port)
	serverConn, _ := net.ListenUDP(
		"udp", 
		&net.UDPAddr{
			IP:[]byte{0,0,0,0},
			Port:10001,
			Zone:"",
			},
	)
	defer serverConn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, _ := serverConn.ReadFromUDP(buf)
		log.Println("Received ", string(buf[0:n]), " from ", addr)
	}
}

//get the location of the file
func whereIs(salt string) (database.Storage, error) {
	return database.Storage{}, errors.New("Not implemented")
}

//Get the best Storage for file
func whereTo(salt string, size int) (database.Storage, error) {
	return database.Storage{}, errors.New("Not implemented")
}

func delete(salt string) error {
	return errors.New("Not implemented")
}

func store(salt string) error {
	return errors.New("Not implemented")
}

//send id back
func subscribe(dns string, used, total uint) (uint, error) {
	return 0, errors.New("Not implemented")
}

func unsubscribe(id uint) error {
	return errors.New("Not implemented")
}