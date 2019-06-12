package server

import (
	"bufio"
	"github.com/gomodule/redigo/redis"
	"loadbalancer/database"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
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

	netData = strings.Trim(netData, "\n")
	args := strings.Split(netData, ArgsDelimiter)
	_, err = strconv.Atoi(args[0])
	if err != nil {
		log.Printf("Error converting string to integer %s ", args[0])
		_, _ = c.Write([]byte(string(UnknownStorage)))
		return
	}
	log.Printf("Original args: %v", args)
	handleRequest(Query(args[0]), args[1:], c)
}


func write(msg Response, c net.Conn) {
	_, err := c.Write([]byte(msg))
	if err != nil {
		log.Printf("Error while writing %v", err)
	}
}

func handleRequest(query Query, args []string, c net.Conn) {
	switch query {
	case SubscribeNew:
		handleSubscribeNew(args, c)
	case SubscribeExisting:
		handleSubscribeExisting(args, c)
	case Unsub:
		handleUnsubscribe(args, c)
	case Store:
		handleStore(args, c)
	case Delete:
		handleDelete(args, c)
	}
}

func handleSubscribeNew(args []string, c net.Conn) {
	arg, err := processSubscribeNew(args)
	if err == nil {
		write(subscribeNew(arg.DNS, arg.UsedSpace, arg.TotalSpace), c)
		return
	}
	log.Printf("Error happened: %v", err)
	write(InternalError, c)
}
func processSubscribeNew(args []string) (Args, error) {
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

func handleSubscribeExisting(args []string, c net.Conn) {
	arg, err := processSubscribeExisting(args)
	if err == nil {
		write(subscribeExisting(arg.ID, arg.DNS, arg.UsedSpace, arg.TotalSpace), c)
		return
	}
	log.Printf("Error happened: %v", err)
	write(InternalError, c)
}

func processSubscribeExisting(args []string) (Args, error) {
	var processed = Args{}
	if len(args) < 4 {
		return processed, NotEnoughArgument(4)
	}
	var err error
	processed.ID, err = stringToUint(args[0])
	if err != nil {
		return processed, ConversionError("ID")
	}
	processed.DNS = args[1]
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

func handleUnsubscribe(args []string, c net.Conn) {
	arg, err := processUnsubscribe(args)
	if err == nil {
		write(unsubscribe(arg.ID), c)
		return
	}
	log.Printf("Error happened: %v", err)
	write(InternalError, c)

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

func handleStore(args []string, c net.Conn) {
	arg, err := processFileOperation(args)
	if err == nil {
		err = store(arg.FileName, c)
		if err != nil {
			log.Printf("Error while storing file: %v", err)
		}
	}
	log.Printf("Error %v", err)
	write(InternalError, c)
}

func processFileOperation(args []string) (Args, error) {
	var processed = Args{}
	if len(args) < 1 {
		return processed, NotEnoughArgument(1)
	}
	processed.FileName = args[0]
	return processed, nil
}

func handleDelete(args []string, c net.Conn) {
	arg, err := processFileOperation(args)
	if err == nil {
		write(deleteFile(arg.FileName), c)
		return
	}
	write(InternalError, c)
}

func stringToUint(id string) (uint, error) {
	result, err := strconv.ParseUint(id, 10, 32)
	return uint(result), err
}