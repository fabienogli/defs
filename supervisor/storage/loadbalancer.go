package storage

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type StoreResp int

const (
	Ok					StoreResp = 0
	HashAlreadyExisting StoreResp = 1
	NoStorageLeft       StoreResp = 2
	HashNotFound        StoreResp = 3
)

func (s StoreResp) String() string {
	return fmt.Sprintf("%d", s)
}

type StoreAction uint8

const (
	WhereTo StoreAction = 0
	WhereIs StoreAction = 1
)

func (a StoreAction) String() string {
	return fmt.Sprintf("%d", int(a))
}

type LoadBalancerClient struct {
	Conn     *net.UDPConn
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
	_ = Conn.SetDeadline(time.Now().Add(time.Second*10))
	_ = Conn.SetReadDeadline(time.Now().Add(time.Second*10))

	if err != nil {
		log.Println("error trying to dial : " +  err.Error())
		//We don't handle the error here because we need to send it back to the client 
		return nil, err

	}

	lb := LoadBalancerClient{
		Conn:     Conn,
	}

	return &lb, nil
}

func (lb *LoadBalancerClient) Close() {
	lb.Conn.Close()
}

func (lb *LoadBalancerClient) Query(code StoreAction, params ... string ) (string, error) {
	query :=  code.String() + " " + strings.Join(params, " ")
	log.Println("querying ", query)
	_, err := lb.Conn.Write([]byte(query))
	if err != nil {
		log.Printf("error writing query %s : %s\n", query, err)
		return "", err
	}

	//Awaiting response
	buf := make([]byte, 1024)
	n, err := lb.Conn.Read(buf)
	if err != nil {
		log.Println("error reading from conn : ", err.Error())
		return "", err
	}

	resp := string(buf[:n])
	return resp, nil
}


func (lb *LoadBalancerClient) WhereTo(hash string, sizeInKb int) (string, error) {
	code := WhereTo
	sizeStr := strconv.Itoa(sizeInKb)
	return lb.Query(code, hash, sizeStr)
}

func (lb *LoadBalancerClient) WhereIs(hash string) (string, error) {
	code := WhereIs
	return lb.Query(code, hash)
}
