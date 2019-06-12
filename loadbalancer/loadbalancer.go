package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"loadbalancer/database"
	"loadbalancer/server"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type Query int

const (
	WhereTo Query = 0
	WhereIs Query = 1
	TTL		int = 60
)

func (q Query) String() string {
	return fmt.Sprintf("%d", q)
}
//@TODO maybe change this to struct to do custom error
type Response int
const (
	OK					Response = 0
	HashAlreadyExisting Response = 1
	NoStorageLeft       Response = 2
	HashNotFound        Response = 3
	MalformRequest      Response = 4
	InternalError      Response = 666
)

func (r Response) String() string {
	return fmt.Sprintf("%d", r)
}

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
		startUdpServer([]byte{0, 0, 0, 0}, port)
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

func startUdpServer(ip []byte, port int) {
	addr := ":" + strconv.Itoa(port)
	log.Printf("Udp server Listening on %s\n", addr)
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

		//If the connection was closed, we return
		if connection == nil {
			return
		}

		//Wait for packets
		n, remoteAddr, err := connection.ReadFromUDP(buffer)

		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				log.Printf("error while reading from socket : %s", err.Error())
				break
			}

			//Timeout error
			log.Printf("timed out")
			continue
		}

		//Parse query from packet
		query := string(buffer[:n])
		log.Println("from", remoteAddr, "-", query)

		queryParts := strings.Split(query, " ")
		if len(queryParts) == 0 {
			//We ignore all malformed queries
			continue
		}

		switch queryParts[0] {
		case WhereIs.String():
			if len(queryParts) != 2 {
				continue
			}

			hash := queryParts[1]
			go HandleWhereIs(connection, remoteAddr, hash)
		case WhereTo.String():
			if len(queryParts) != 3 {
				continue
			}

			hash := queryParts[1]
			size, err := strconv.Atoi(queryParts[2])
			if err != nil {
				continue
			}
			go HandleWhereTo(connection, remoteAddr, hash, size)
		default:
			//Ignore by default
			continue
		}
	}
	quit <- true
}

func HandleWhereTo(connection *net.UDPConn, addr *net.UDPAddr, hash string, size int) {
	//todo Query DB here with function whereTo(hash, size) etc...
	response := MalformRequest.String()
	if size < 0 {
		Respond(connection, addr, response)
		return
	}
	_, err := database.GetFile(hash, conn)
	if err == nil {
		storage, code := getLargestStorage(uint(size))
		if code == OK {
			file, err := database.NewFile(hash, storage.ID, size)
			erratum := tempStore(file)
			if err == nil && erratum == nil {
				response = fmt.Sprintf("%d %s", OK, storage.DNS)
			} else {
				response = MalformRequest.String()
				log.Printf("Error while saving file %v\nor temp Storing %v", err, erratum)
			}
		} else {
			response = code.String()
		}
	} else {
		response = HashAlreadyExisting.String()
	}
	Respond(connection, addr, response)
}

func HandleWhereIs(connection *net.UDPConn, addr *net.UDPAddr, hash string) {
	// TODO irindul 2019-05-28 : Query DB here with function whereIs(hash)

	resp := OK.String() + " storage"
	Respond(connection, addr, resp)
}

func Respond(connection *net.UDPConn, addr *net.UDPAddr, resp string) {
	respBytes := []byte(resp)
	_, err := connection.WriteToUDP(respBytes, addr)
	if err != nil {
		log.Println("could not write response ", err.Error())
	}
}

//get the location of the file
func whereIs(hash string) (database.Storage, error) {
	file, err := database.GetFile(hash, conn)
	if err == nil {
		storage, err := database.GetStorage(file.DNS, conn)
		return storage, err
	}
	return database.Storage{}, err
}

//Get the best Storage for file
func whereTo(hash string, size int) Response {
	_, err := database.GetFile(hash, conn)
	if err != nil {
		return HashAlreadyExisting
	}
	if size < 0 {
		return MalformRequest
	}
	storage, resp := getLargestStorage(uint(size))
	file, err := database.NewFile(hash, storage.ID, size)
	err = tempStore(file)
	if err != nil {
		return MalformRequest
	}
	return fmt.Sprintf("%d %s", OK, storage.DNS)
}

func getLargestStorage(size uint) (database.Storage, Response) {
	storages, _ := database.GetStorages(conn)
	var largest uint
	var i = -1
	for index, storage := range storages {
		temp := storage.GetAvailableSpace()
		if temp < size {
			continue
		}
		if largest < temp {
			largest = temp
			i = index
		}
	}
	if i >= 0 {
		return storages[i], OK
	}
	return database.Storage{}, NoStorageLeft
}

func tempStore(file database.File) error {
	err := file.Create(conn)
	if err != nil {
		return err
	}
	err = file.SetExp(TTL, conn)
	return err
}