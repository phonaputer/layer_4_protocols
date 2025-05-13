package main

import (
	"fmt"
	"net"
	"udp/server"
)

func main() {
	conn, err := server.ListenUDP("udp4", &server.Address{
		IP:   &net.IPAddr{IP: []byte{127, 0, 0, 1}},
		Port: 8080,
	})
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		panic(err)
	}

	fmt.Println("got message from: ", addr.IP.String(), addr.Port)
	fmt.Println(string(buf[:n]))
}
