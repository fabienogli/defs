package main

import (
	"net"
	"log"
	"loadbalancer/database"
	"errors"
	"os"
	"runtime"
	"strconv"
)

func main() {
	port := os.Getenv("port")
	if port == "" {
		panic("Port wasn't found\n")
	}
	i, err := strconv.Atoi(port)
	if err != nil {
		log.Println(err)
		panic("Error converting port from string to int")
	}
	addr := ":" + port
	log.Printf("Listening on %s\n", addr)
	serverConn, err := net.ListenUDP(
		"udp", 
		&net.UDPAddr{
			IP:[]byte{0,0,0,0},
			Port:i,
			Zone:"",
			},
	)
	if err != nil {

	}
	defer serverConn.Close()
	quit := make(chan struct{})
	for i := 0; i < runtime.NumCPU(); i++ {
		go listen(serverConn, quit)
	}
	<- quit
}

/*
func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
    _,err := conn.WriteToUDP([]byte("From server: Hello I got your mesage "), addr)
    if err != nil {
        fmt.Printf("Couldn't send response %v", err)
    }
}


func main() {
    p := make([]byte, 2048)
    addr := net.UDPAddr{
        Port: 1234,
        IP: net.ParseIP("127.0.0.1"),
    }
    ser, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Printf("Some error %v\n", err)
        return
    }
    for {
        _,remoteaddr,err := ser.ReadFromUDP(p)
        fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
        if err !=  nil {
            fmt.Printf("Some error  %v", err)
            continue
        }
        go sendResponse(ser, remoteaddr)
    }
}
 */

func listen(connection *net.UDPConn, quit chan struct{}) {
	buffer := make([]byte, 1024)
	n, remoteAddr, err := 0, new(net.UDPAddr), error(nil)
	for err == nil {
			n, remoteAddr, err = connection.ReadFromUDP(buffer)
			log.Println("from", remoteAddr, "-", string(buffer[:n]))
			_, err := connection.WriteToUDP([]byte("storage"), remoteAddr)
			if err != nil {
				log.Println(err.Error())
			}
	}
	log.Println("listener failed - ", err)
	quit <- struct{}{}
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