package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type LoadBalancerCode uint8

const (
	SubscribeNew      LoadBalancerCode = iota
	SubscribeExisting LoadBalancerCode = iota
	Unsub             LoadBalancerCode = iota
	Store             LoadBalancerCode = iota
	Delete            LoadBalancerCode = iota
	DoneStoring       LoadBalancerCode = iota
)

type ResponseCode uint64

const (
	Ok                 ResponseCode = iota
	StorageNonExistent ResponseCode = 1
	NotSameUsedSpace   ResponseCode = 2
	UnknownStorage     ResponseCode = 3
	UnknownFile        ResponseCode = 4
	InternalError      ResponseCode = 666
)

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

func Subscribe() {
	id := GetId()
	addr := GetTCPAddr()
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
	}
	if id == "" {
		SubscibeWithoutId(conn)
	} else {
		SubscribeWithId(id, conn)
	}
}

func SubscribeWithId(id string, conn net.Conn) {
	myDns := os.Getenv("STORAGE_DNS")
	totalSpace := os.Getenv("STORAGE_SPACE")
	usedSpace := fmt.Sprintf("%d", GetUsedSpace())

	query := craftQuery(SubscribeExisting, id, myDns, usedSpace, totalSpace)

	buf := []byte(query)
	n, err := conn.Write(buf)
	if err != nil {
		log.Panicf("error writing bytes to conn : %s", err)
	}
	if n != len(buf) {
		log.Panicf("error wrote %d bytes but should have written %d", n, len(buf))
	}

	buf = make([]byte, 2048)
	n, err = conn.Read(buf)
	if err != nil {
		log.Panicf("could not read from connection : %s", err)
	}
	response := string(buf[:n])

	HandleExisting(response)
}

func HandleExisting(response string) {
	responseParts := strings.Split(response, " ")
	status, err := strconv.Atoi(responseParts[0])
	if err != nil {
		log.Panicf("could not convert response into int : %s", err)
	}
	switch ResponseCode(status) {
	case Ok:
		//We are all set baby ! We don't need any operations
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

func SubscibeWithoutId(conn net.Conn) {
	//defer conn.Close()
	myDns := os.Getenv("STORAGE_DNS")
	totalSpace := os.Getenv("STORAGE_SPACE")
	usedSpace := fmt.Sprintf("%d", GetUsedSpace())

	query := craftQuery(SubscribeNew, myDns, usedSpace, totalSpace)

	buf := []byte(query)
	n, err := conn.Write(buf)
	if err != nil {
		log.Panicf("error writing bytes to conn : %s", err)
	}

	if n != len(buf) {
		log.Panicf("error wrote %d bytes but should have written %d", n, len(buf))
	}

	//We make a buffer big enough for the answer, 2048 is way overkill but at least we are sure
	buf = make([]byte, 2048)
	n, err = conn.Read(buf)
	if err != nil {
		log.Panicf("could not read from connection : %s", err)
	}
	response := string(buf[:n])

	HandleNewId(response)

}

func HandleNewId(response string) {
	responseParts := strings.Split(response, " ")
	status, err := strconv.Atoi(responseParts[0])
	if err != nil {
		log.Panicf("could not convert response into string : %s", err)
	}
	switch ResponseCode(status) {
	case Ok:
		id := responseParts[1]
		createIdFile(id)
	default:
		log.Panicf("bad response from loadbalancer : %s", response)
	}

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

func GetUsedSpace() int64 {
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

func craftQuery(code LoadBalancerCode, args ...string) string {
	return fmt.Sprintf("%d %s\n", code, strings.Join(args, " "))
}
