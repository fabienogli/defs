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

type LoadBalancerCode uint8

const (
	SubscribeNew   LoadBalancerCode = iota
	SubscribeExist LoadBalancerCode = iota
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
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Panicf("could not dial loadbalancer : %s", err)
	}


	if id == "" {
		SubscibeWithoutId(conn)
	} else {
		SubscribeWithId(id, conn)
	}
}

func SubscribeWithId(id string, conn net.Conn) {

}

func SubscibeWithoutId(conn net.Conn) {
	defer conn.Close()
	myDns := os.Getenv("STORAGE_DNS")
	totalSpace := os.Getenv("STORAGE_SPACE")
	usedSpace := fmt.Sprintf("%d", GetUsedSpace())
	log.Println(myDns)
	log.Println(totalSpace)
	query := writeQuery(SubscribeNew, myDns, usedSpace, totalSpace)

	buf := []byte(query)
	n, err := conn.Write(buf)
	if err != nil {
		log.Panicf("error writing bytes to conn : %s", err)
	}

	if n != len(buf) {
		log.Panicf("error wrote %d bytes but should have written %d", n, len(buf))
	}

	//Wait for response

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

func writeQuery(code LoadBalancerCode, args ...string) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%d ", code))
	sb.WriteString(strings.Join(args, " "))
	return sb.String()
}
