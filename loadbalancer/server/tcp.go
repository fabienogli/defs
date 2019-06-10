package server

import (
	"bufio"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"loadbalancer/database"
	"errors"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type Query string
type Response string

const (
	SubscribeNew      Query = "0"
	SubscribeExisting Query = "1"
	Unsub             Query = "2"
	Store             Query = "3"
	Delete            Query = "4"

	Ok                 Response = "0"
	StorageNonExistent Response = "1"
	NotSameUsedSpace   Response = "2"
	UnknownStorage     Response = "2"
	InternalError      Response = "666"

	CmdDelimiter byte = '\n'
	ArgsDelimiter string = " "

)

var conn redis.Conn

var ErrorConversion error

func StartTCP(port int) error {
	pool := database.GetDatabase()
	conn = pool.Get()
	listening := ":" + strconv.Itoa(port)
	l, err := net.Listen("tcp4", listening)
	if err != nil {
		return err
	}
	log.Printf("Tcp Server listening on %s\n", listening)
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConnection(c)
	}
}


func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	netData, err := bufio.NewReader(c).ReadString(CmdDelimiter)
	if err != nil {
		fmt.Println(err)
		return
	}

	//process cmd
	args := strings.Split(netData, ArgsDelimiter)
	code, err := strconv.Atoi(args[0])
	if err != nil {
		log.Printf("Error converting string to integer %s ", args[0])
		_, _ = c.Write([]byte(string(UnknownStorage)))
	}
	query := Query(code)
	handleRequest(query, args[1:])
	_ = c.Close()
}

func handleRequest(query Query, args []string) string {
	var resp string = string(InternalError)
	switch query {
	case SubscribeNew:
		totalSpace, err := stringToUint(args[2])
		if err != nil {
			return resp
		}
		UsedSpace, err := stringToUint(args[1])
		if err != nil {
			return resp
		}
		id, response := subscribeNew(args[0], UsedSpace, totalSpace)
		stringId := uintToString(id)
		resp = string(response)+ ArgsDelimiter + stringId
		break
	case SubscribeExisting:
		UsedSpace, err := stringToUint(args[2])
		if err != nil {
			return resp
		}
		totalSpace, err := stringToUint(args[3])
		if err != nil {
			return resp
		}
		id, err := stringToUint(args[0])
		if err == nil {
			response := subscribeExisting(id, args[0], UsedSpace, totalSpace)
			resp = string(response)
		}
		break
	case Unsub:
		id, err := stringToUint(args[0])
		if err != nil {
			return string(InternalError)
		}
		err = unsubscribe(id)
		if err == nil {
			resp = string(Ok)
		}
		break
	case Store:
		err := store(args[0])
		if err == nil {
			resp = string(Ok)
		}
	case Delete:
		err := delete(args[0])
		if err == nil {
			resp = string(Ok)
		}
	}
	return resp
}

func stringToUint(id string) (uint, error) {
	result, err := strconv.ParseUint(id, 10, 32)
	return uint(result), err
}

func uintToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
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
func subscribeNew(dns string, used, total uint) (uint, Response) {
	storage := database.Storage{
		DNS:   dns,
		Used:  used,
		Total: total,
	}
	storage.GenerateUid(conn)
	err := storage.Create(conn)
	if err != nil {
		return 0, InternalError
	}
	return storage.ID, Ok
}

func subscribeExisting(id uint, dns string, used, total uint) Response {
	storage := database.Storage{
		ID:    id,
		DNS:   dns,
		Used:  used,
		Total: total,
	}
	dbStorage, err := database.GetStorage(id, conn)
	if err != nil {
		return StorageNonExistent
	}
	if storage.Used != dbStorage.Used {
		return NotSameUsedSpace
	}
	err = storage.Update(conn)
	if err != nil {
		return InternalError
	}
	return Ok
}

func unsubscribe(id uint) error {
	storage, err := database.GetStorage(id, conn)
	if err != nil {
		return err
	}
	err = storage.Delete(conn)
	return err
}
