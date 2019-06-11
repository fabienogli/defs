package server

import (
	"bufio"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"loadbalancer/database"
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
	DoneStoring       Query = "5"
	TTL               int   = 60

	Ok                 Response = "0"
	StorageNonExistent Response = "1"
	NotSameUsedSpace   Response = "2"
	UnknownStorage     Response = "3"
	UnknownFile        Response = "4"
	InternalError      Response = "666"

	CmdDelimiter  byte   = '\n'
	ArgsDelimiter string = " "
)

type Args struct {
	ID uint
	FileName string
	StoreName string
	DNS string
	UsedSpace uint
	TotalSpace uint
}

var conn redis.Conn

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
	defer c.Close()

	//Read data
	netData, err := bufio.NewReader(c).ReadString(CmdDelimiter)
	if err != nil {
		log.Printf("Error while reading %v", err)
		return
	}

	//Process cmd
	netData = strings.Trim(netData, "\n")
	args := strings.Split(netData, ArgsDelimiter)
	_, err = strconv.Atoi(args[0])
	if err != nil {
		log.Printf("Error converting string to integer %s ", args[0])
		_, _ = c.Write([]byte(string(UnknownStorage)))
		return
	}
	resp := handleRequest(Query(args[0]), args[1:], c)
	_, err = c.Write([]byte(resp))
	if err != nil {
		log.Printf("Error while writing response %v", err)
	}
}

type NotEnoughArgument uint
type ConversionError string

func (n NotEnoughArgument) Error() string {
	return fmt.Sprintf("Not Enough Argument : %d", n)
}
func (c ConversionError) Error() string {
	return fmt.Sprintf("Error converting %s", c)
}

func processSubcribeNew(args []string) (Args, error) {
	var processed = Args{}
	if len(args) < 3 {
		return processed, NotEnoughArgument(3)
	}
	processed.DNS = args[0]
	var err error
	processed.TotalSpace, err = stringToUint(args[2])
	if err != nil {
		return processed, ConversionError("Total Space")
	}
	processed.UsedSpace, err = stringToUint(args[1])
	if err != nil {
		return processed, ConversionError("Used Space")
	}
	return processed, nil
}

func processSubcribeExisting(args []string) (Args, error) {
	var processed = Args{}
	if len(args) < 4 {
		return processed, NotEnoughArgument(4)
	}
	id, err := stringToUint(args[0])
	if err != nil {
		return processed, ConversionError("ID")
	}
	dns := args[1]
	processed.UsedSpace, err = stringToUint(args[2])
	if err != nil {
		return processed, ConversionError("USED Space")
	}
	processed.TotalSpace, err = stringToUint(args[3])
	if err != nil {
		return processed, ConversionError("Total Space")
	}
	return processed, nil
}

func processUnsubscribe(args []string) (Args, error) {
	var processed = Args{}
	id, err := stringToUint(args[0])
	if err != nil {
		return processed, ConversionError("ID")
	}
	processed.ID = id
	return processed, nil
}

func processFileOperation(args []string) (Args, error) {
	var processed = Args{}
	if len(args) < 1 {
		return processed, NotEnoughArgument(1)
	}
	processed.FileName = args[0]
	return processed, nil
}


func handleRequest(query Query, args []string, c net.Conn) string {
	var resp = string(InternalError)
	switch query {
	case SubscribeNew:
		//process args
		if len(args) < 3 {
			return resp
		}
		dns := args[0]
		totalSpace, err := stringToUint(args[2])
		if err != nil {
			log.Printf("Error: %v", err)
			return resp
		}
		UsedSpace, err := stringToUint(args[1])
		if err != nil {
			return resp
		}
		id, response := subscribeNew(dns, UsedSpace, totalSpace)
		resp = fmt.Sprintf("%s %s %d", response, ArgsDelimiter, id)
	case SubscribeExisting:
		if len(args) < 4 {
			return resp
		}
		id, err := stringToUint(args[0])
		dns := args[1]
		UsedSpace, err := stringToUint(args[2])
		if err != nil {
			return resp
		}
		totalSpace, err := stringToUint(args[3])
		if err == nil {
			resp = fmt.Sprintf("%s", subscribeExisting(id, dns, UsedSpace, totalSpace))
		}
	case Unsub:
		id, err := stringToUint(args[0])
		if err != nil {
			return string(InternalError)
		}
		resp = string(unsubscribe(id))
	case Store:
		if len(args) < 1 {
			log.Printf("WTF")
			return resp
		}
		fileName := args[0]
		if err := store(fileName, c); err == nil {
			resp = string(Ok)
		} else {
			log.Printf("Error while storing: %v", err)
		}
	case Delete:
		if len(args) < 1 {
			return resp
		}
		filename := args[0]
		resp = string(delete(filename))
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

func delete(hash string) Response {
	file, err := database.GetFile(hash, conn)
	if err != nil {
		return UnknownFile
	}
	err = file.Delete(conn)
	if err == nil {
		return Ok
	}
	return InternalError
}

func store(hash string, c net.Conn) error {
	file, err := database.GetFile(hash, conn)
	if err != nil {
		return err
	}

	err = file.Persist(conn)

	if err != nil {
		return err
	}

	//Read data
	netData, err := bufio.NewReader(c).ReadString(CmdDelimiter)
	args := strings.Split(netData, ArgsDelimiter)
	if err == nil && len(args) >= 2 {
		code, err := strconv.Atoi(args[0])
		if Query(code) == DoneStoring && err == nil {
			return nil
		} else {
			log.Printf("Unknown code")
		}
	} else {
		log.Printf("Error while reading file : %v\nArguments: %v", err, args)
	}
	err = file.SetExp(TTL, conn)
	return err
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
		log.Printf("Error while creating storage: %v\nstorage: %v", err, storage)
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

func unsubscribe(id uint) Response {
	storage, err := database.GetStorage(id, conn)
	if err != nil {
		return UnknownStorage
	}
	err = storage.Delete(conn)
	if err == nil {
		return Ok
	}
	return InternalError
}
