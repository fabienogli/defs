package storage

import (
	"fmt"
	"log"
	"net"
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

func NewLoadBalancerClient(ip string, port int) (*LoadBalancerClient, error) {
	ipAddr := net.ParseIP(ip)
	Conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   ipAddr,
		Port: port,
		Zone: "",
	})

	if err != nil {
		//We don't handle the error here because we need to send it back to the client
		// TODO irindul 2019-05-20 : Check if its timed out, maybe retry a couple of times
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
		// TODO irindul 2019-05-20 : Handle Connection closed etc...
	}
	buf := make([]byte, 1024)
	n, err := lb.Conn.Read(buf)
	if err != nil {
		// TODO irindul 2019-05-20 : Handle error
	}
	log.Printf("Dns received : %s\n", string(buf[:n]))
}
