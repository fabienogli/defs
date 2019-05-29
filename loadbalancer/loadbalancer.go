package main

import (
	"errors"
	"fmt"
	"loadbalancer/database"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	portStr := os.Getenv("LOADBALANCER_PORT")
	if portStr == "" {
		panic("Port wasn't found\n")
	}
	port, err:= strconv.Atoi(portStr)
	if err != nil {
		log.Println(err)
		panic(fmt.Sprintf("error converting port from string to int : %S", err.Error()))
	}
	startUdpServer([]byte{0,0,0,0}, port)
}

func startUdpServer(ip []byte, port int) {
	addr := ":" + strconv.Itoa(port)
	log.Printf("Listening on %s\n", addr)
	serverConn, _ := net.ListenUDP(
		"udp",
		&net.UDPAddr{
			IP:   ip,
			Port: port,
			Zone: "",
		},
	)
	defer serverConn.Close()
	quit := make(chan bool)

	//Run as many listener as possible
	for i := 0; i < runtime.NumCPU(); i++ {
		go listen(serverConn, quit)
	}
	<-quit
	log.Println("quiting server")
}

func listen(connection *net.UDPConn, quit chan bool) {
	buffer := make([]byte, 1024)

	for {
		//Wait for packets
		n, remoteAddr, err := connection.ReadFromUDP(buffer)

		if err != nil {
			// TODO irindul 2019-05-28 : Handle ConnectionTimedOut etc...
			log.Printf("error while reading from socket : %s", err.Error())
			break
		}

		//Parse query from packet
		query := string(buffer[:n])
		log.Println("from", remoteAddr, "-", query)

		queryPart := strings.Split(query, " ")

		// TODO irindul 2019-05-28 : Put enum here
		// TODO irindul 2019-05-28 : Handle queryPart malformed here!
		switch queryPart[0] {
		case "1":
			hash := queryPart[1]
			go HandleWhereIs(connection, remoteAddr, hash)
		case "0":
			hash := queryPart[1]
			size, _ := strconv.Atoi(queryPart[2]) //todo check error (malformed request)
			go HandleWhereTo(connection, remoteAddr, hash, size)
		//HandleWhereTo()
		case "":
		default:
			//Ignore
			//Handle default case
		}
	}
	quit <- true
}

func HandleWhereTo(connection *net.UDPConn, addr *net.UDPAddr, hash string, size int) {
	//todo Query DB here with function whereTo(hash, size) etc...
	//todo craft response with enum

	resp := "0 storage"
	Respond(connection, addr, resp)


}

func HandleWhereIs(connection *net.UDPConn, addr *net.UDPAddr, hash string) {

	// TODO irindul 2019-05-28 : Query DB here with function whereIs(hash)

	//todo craft response with enum
	resp := "0 storage"
	Respond(connection, addr, resp)
}

func Respond(connection *net.UDPConn, addr *net.UDPAddr, resp string) {
	respBytes := []byte(resp)
	log.Println("Paylaod size ")
	_, err := connection.WriteToUDP(respBytes, addr)
	if err != nil {
		log.Println("could not write response ", err.Error())
	}
}

//get the location of the file
func whereIs(hash string) (database.Storage, error) {
	return database.Storage{}, errors.New("Not implemented")
}

//Get the best Storage for file
func whereTo(hash string, size int) (database.Storage, error) {
	return database.Storage{}, errors.New("Not implemented")
}

func delete(hash string) error {
	return errors.New("Not implemented")
}

func store(hash string) error {
	return errors.New("Not implemented")
}

//send id back
func subscribe(dns string, used, total uint) (uint, error) {
	return 0, errors.New("Not implemented")
}

func unsubscribe(id uint) error {
	return errors.New("Not implemented")
}
