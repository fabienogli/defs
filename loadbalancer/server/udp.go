package server

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"net"
	"runtime"
	"loadbalancer/database"
	"strconv"
	"strings"
)

func StartUDP(ip string, port int) {
	ipParsed := net.ParseIP(ip)
	addr := ":" + strconv.Itoa(port)
	log.Printf("Udp server Listening on %s\n", addr)
	serverConn, _ := net.ListenUDP(
		"udp",
		&net.UDPAddr{
			IP:   ipParsed,
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
		query = strings.Trim(query, "\n")
		log.Println("from", remoteAddr, "-", query)

		queryParts := strings.Split(query, " ")
		if len(queryParts) == 0 {
			//We ignore all malformed queries
			continue
		}

		switch Query(queryParts[0]) {
		case WhereIs:
			if len(queryParts) != 2 {
				continue
			}

			hash := queryParts[1]
			HandleWhereIs(connection, remoteAddr, hash)
		case WhereTo:
			if len(queryParts) != 3 {
				continue
			}

			hash := queryParts[1]
			size, err := strconv.Atoi(queryParts[2])
			if err == nil {
				HandleWhereTo(connection, remoteAddr, hash, size)
				continue
			}
			log.Printf("Error happened %v", err)
		default:
			//Ignore by default
			continue
		}
	}
	quit <- true
}

func HandleWhereTo(connection *net.UDPConn, addr *net.UDPAddr, hash string, size int) {
	response := MalformRequest.String()
	if size < 0 {
		Respond(connection, addr, response)
		return
	}
	_, err := database.GetFile(hash, conn)
	if err != redis.ErrNil {
		log.Printf("Error happened %v", err)
		response = HashAlreadyExisting.String()
		Respond(connection, addr, response)
		return
	}
	storage, code := getLargestStorage(uint(size))
	if code == OK {
		file, err := database.NewFile(hash, storage.ID, size)
		erratum := tempStore(file)
		if err == nil && erratum == nil {
			response = fmt.Sprintf("%s %s", OK, storage.DNS)
		} else {
			response = MalformRequest.String()
			log.Printf("Error while saving file %v\nor temp Storing %v", err, erratum)
		}
	} else {
		response = code.String()
	}
	Respond(connection, addr, response)
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

func HandleWhereIs(connection *net.UDPConn, addr *net.UDPAddr, hash string) {
	resp := HashNotFound.String()
	file, err := database.GetFile(hash, conn)
	if err == nil {
		storage, err := database.GetStorage(file.DNS, conn)
		if err == nil { resp = fmt.Sprintf("%s %s", OK, storage.DNS)}
	}
	Respond(connection, addr, resp)
}

func Respond(connection *net.UDPConn, addr *net.UDPAddr, resp string) {
	respBytes := []byte(resp)
	_, err := connection.WriteToUDP(respBytes, addr)
	if err != nil {
		log.Println("could not write response ", err.Error())
	}
}

func tempStore(file database.File) error {
	err := file.Create(conn)
	if err == nil {
		err = file.SetExp(TTL, conn)
	}
	return err
}