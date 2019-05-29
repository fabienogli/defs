package storage

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

type StoreErr uint8

const (
	HashAlreadyExisting StoreErr = 1
	NoStorageLeft       StoreErr = 2
	HashNotFound        StoreErr = 3
)

type StoreAction uint8

const (
	WhereTo StoreAction = 0
	WhereIs StoreAction = 1
)

type LoadBalancerClient struct {
	Conn     *net.UDPConn
	Messages chan string
}

func GetUDPAddrOfLB() (*net.UDPAddr,error) {
	host := os.Getenv("LOADBALANCER_IP")
	portStr := os.Getenv("LOADBALANCER_PORT")

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{
		IP: ips[0],
		Port: port,
		Zone: "",
	}, nil
}

func NewLoadBalancerClient() (*LoadBalancerClient, error) {

	udpAddr, err := GetUDPAddrOfLB()
	if err != nil {
		log.Println("could not get UDP address of loadbalancer", err)
		return nil, err
	}

	Conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		log.Println("error trying to dial : " +  err.Error())
		//We don't handle the error here because we need to send it back to the client 
		return nil, err
	}

	lb := LoadBalancerClient{
		Conn:     Conn,
		Messages: make(chan string),
	}

	return &lb, nil
}

func (lb *LoadBalancerClient) Close() {
	lb.Conn.Close()
	close(lb.Messages)
}

func (lb *LoadBalancerClient) WhereTo(hash string, sizeInKb int) {
	code := WhereTo
	query := fmt.Sprintf("%d %s %d", code, hash, sizeInKb)
	_, err := lb.Conn.Write([]byte(query))
	if err != nil {
		log.Println("error : ", err.Error())
	}


	buf := make([]byte, 1024)
	n, err := lb.Conn.Read(buf)
	if err != nil {
		log.Println("error : ", err.Error())
	}

	log.Printf("Dns received : %s\n", string(buf[:n]))
}
