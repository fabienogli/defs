package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type QueryCode uint8

const (
	SubscribeNew      QueryCode = iota
	SubscribeExisting QueryCode = iota
	Unsub             QueryCode = iota
	StoreStart        QueryCode = iota
	Delete            QueryCode = iota
	StoreDone         QueryCode = iota
)

type ResponseCode string

const (
	Ok                 ResponseCode = "0"
	StorageNonExistent ResponseCode = "1"
	NotSameUsedSpace   ResponseCode = "2"
	UnknownStorage     ResponseCode = "3"
	UnknownFile        ResponseCode = "4"
	InternalError      ResponseCode = "666"
)

type Args struct {
	id         string
	dns        string
	usedSpace  string
	totalSpace string
}

func GetTCPAddr() string {
	ip := os.Getenv("LOADBALANCER_IP")
	port := os.Getenv("LOADBALANCER_PORT")

	return ip + ":" + port
}

func GetId() string {
	idPath := os.Getenv("STORAGE_ID_FILE")
	idFile, err := os.Open(idPath)
	if err != nil {
		//All errors means that I don't have an id
		return ""
	}
	defer idFile.Close()

	//We make a buffer big enough
	buf := make([]byte, 256)
	n, err := idFile.Read(buf)
	if err != nil && err != io.EOF {
		//Huston, we have a problem
		log.Panicf("error reading id : %s", err)
		return ""
	}

	return string(buf[:n])
}

func ConnectToLoadBalancer() net.Conn {
	addr := GetTCPAddr()
	for {
		conn, err := net.Dial("tcp", addr)

		if err != nil {
			continue
		}
		return conn
	}
}

func Subscribe() {
	id := GetId()
	conn := ConnectToLoadBalancer()
	defer conn.Close()

	dns := os.Getenv("STORAGE_DNS")
	if dns == "" {
		log.Panicf("The STORAGE_DNS env variable must be set")
	}

	totalSpace := os.Getenv("STORAGE_SPACE")
	if totalSpace == "" {
		log.Panicf("The STORAGE_SPACE env variable must be set")
	}

	usedSpace := fmt.Sprintf("%d", getUsedSpace())

	args := Args{
		id:         id,
		dns:        dns,
		usedSpace:  usedSpace,
		totalSpace: totalSpace,
	}

	sendSubscription(args, conn)
}

func Unsubscribe() {
	conn := ConnectToLoadBalancer()
	defer conn.Close()

	id := GetId()
	query := craftQuery(Unsub, id)

	writeQueryToConn(query, conn)
	response := readResponse(conn)
	responseParts := strings.Split(response, " ")
	switch ResponseCode(responseParts[0]) {
	case Ok:
	//Subscribe went well
	default:
		log.Panicf("There was a problem unsubsribing... %s : ", response)
	}
}

func Store(done chan bool, errs chan error, filename string) {
	conn := ConnectToLoadBalancer()
	query := craftQuery(StoreStart, filename)
	writeQueryToConn(query, conn)

	select {
	case <-done:
		query = craftQuery(StoreDone, filename)
		writeQueryToConn(query, conn)
	case <-errs:
		conn.Close()
		return
	}

	conn.Close()
}

func sendSubscription(args Args, conn net.Conn) {
	var code QueryCode
	var query string
	isNewSubscription := args.id == ""

	if isNewSubscription {
		code = SubscribeNew
		query = craftQuery(code, args.dns, args.usedSpace, args.totalSpace)
	} else {
		code = SubscribeExisting
		query = craftQuery(code, args.id, args.dns, args.usedSpace, args.totalSpace)
	}

	writeQueryToConn(query, conn)
	response := readResponse(conn)
	responseParts := strings.Split(response, " ")
	switch ResponseCode(responseParts[0]) {
	case Ok:
		if isNewSubscription {
			if len(responseParts) != 2 {
				log.Panicf("bad response from loadbalancer : %s", response)
			}
			id := responseParts[1]
			createIdFile(id)
		}
	case NotSameUsedSpace:
		//This is where we would handle rollback options => An exchange would start between the loadbalancer and the storage
		//Each files currently stored on the storage would be sent to the loadbalancer
		//The loadbalancer must compare it to the file list it have, and get up to date. (Delete old file, add new files etc...)
		//By lack of time, this feature is not implemented right now and the program will just panic
		log.Panicf("storage out of sync with loadbalancer, please check file lists manually")
	case StorageNonExistent:
		log.Panicf("id is corrupted, the loadbalancer doesn't know me :'(")
	case InternalError:
		log.Panicf("the loadbalancer crashed for some obscure reason")
	default:
		log.Panicf("bad response from loadbalancer : %s", response)
	}
}

func writeQueryToConn(query string, conn net.Conn) {
	buf := []byte(query)
	n, err := conn.Write(buf)
	if err != nil {
		log.Panicf("error writing bytes to conn : %s", err)
	}
	if n != len(buf) {
		log.Panicf("error wrote %d bytes but should have written %d", n, len(buf))
	}
}

func readResponse(conn net.Conn) string {
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		log.Panicf("could not read from connection : %s", err)
	}
	response := string(buf[:n])
	return response
}

func createIdFile(id string) {
	idPath := os.Getenv("STORAGE_ID_FILE")
	if idPath == "" {
		log.Panicf("no STORAGE_ID_FILE was set, please make sure to export this env variable")
	}
	idFile, err := os.OpenFile(idPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Panicf("could not open %s : %s", idPath, err)
	}
	defer idFile.Close()
	r := strings.NewReader(id)
	n, err := io.Copy(idFile, r)
	if err != nil {
		log.Panicf("could not write id to %s : %s", idPath, err)
	}
	if n != int64(len(id)) {
		log.Panicf("partial write of id, %d was written instead of %d", n, len(id))
	}
}

func getUsedSpace() int64 {
	path := os.Getenv("STORAGE_DIR")
	var size int64
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return 0
	}
	err := filepath.Walk(os.Getenv("STORAGE_DIR"), func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	if err != nil {
		log.Panicf("could not read used space : %s", err)
	}

	return size
}

func craftQuery(code QueryCode, args ...string) string {
	return fmt.Sprintf("%d %s\n", code, strings.Join(args, " "))
}
