package main

import(
	"net"
)

func client() {
	Conn, _ := net.DialUDP("udp", nil, &net.UDPAddr{IP:[]byte{127,0,0,1},Port:8081,Zone:""})
	defer Conn.Close()
	Conn.Write([]byte("hello"))
}