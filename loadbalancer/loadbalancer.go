package main

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"defs/loadbalancer/database"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
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
)

func (r Response) String() string {
	return fmt.Sprintf("%d", r)
}

var conn redis.Conn

func getDatabase() *redis.Pool{
	addr := os.Getenv("ROUTING_DB_ADDR")
	sport := os.Getenv("ROUTING_DB_PORT")
	port,_  := strconv.Atoi(sport)
	return database.NewPool(addr, port)
}

func main() {
	portStr := os.Getenv("LOADBALANCER_PORT")
	pool := getDatabase()
	conn := pool.Get()
	defer conn.Close()
	if portStr == "" {
		panic("Port wasn't found\n")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err)
		panic(fmt.Sprintf("error converting port from string to int : %s", err.Error()))
	}
	startUdpServer([]byte{0, 0, 0, 0}, port)
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

	resp := OK.String() + " storage"
	Respond(connection, addr, resp)
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
func whereTo(hash string, size int) (database.Storage, Response) {
	_, err := database.GetFile(hash, conn)
	if err != nil {
		return database.Storage{}, HashAlreadyExisting
	}
	if size < 0 {
		return database.Storage{}, MalformRequest
	}
	storage, resp := getLargestStorage(uint(size))
	file, err := database.NewFile(hash, storage.ID, size)
	err = tempStore(file)
	if err != nil {
		return database.Storage{}, MalformRequest
	}
	return storage, resp  
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

func delete(hash string) error {
	file, err := database.GetFile(hash, conn)
	if err != nil {
		return err
	}
	err = file.Delete(conn)
	return err
}

func store(hash string) error {
	file, err := database.GetFile(hash, conn)
	if err != nil {
		return err
	}

	//@TODO when tcp server operationnal: wait for connection
	err = file.Persist(conn)
	if err != nil {
		return err
	}
	//@TODO if connection fail, delete file

	return errors.New("Not implemented")
}

//@TODO shouldn't be good to delete all file as well ?
func subscribe(dns string, used, total uint) (uint, error) {
	storage := database.Storage{
		DNS:dns,
		Used:used,
		Total:total,
	}
	storage.GenerateUid(conn)
	err := storage.Save(conn)
	return storage.ID, err
}

func unsubscribe(id uint) error {
	storage, err := database.GetStorage(id, conn)
	if err != nil {
		return err
	}
	err = storage.Delete(conn)
	return err
}
