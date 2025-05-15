package main

import (
	"fmt"
	"net"
	"sync"
	"udp/udp"
)

func main() {
	serverWg := sync.WaitGroup{}
	clientWg := sync.WaitGroup{}

	srv, err := udp.Listen("udp4", &udp.Address{
		IP:   &net.IPAddr{IP: []byte{127, 0, 0, 1}},
		Port: 8080,
	})
	if err != nil {
		panic(err)
	}

	cli, err := udp.Listen("udp4", &udp.Address{
		IP:   &net.IPAddr{IP: []byte{127, 0, 0, 1}},
		Port: 8081,
	})
	if err != nil {
		panic(err)
	}

	serverWg.Add(1)
	go server(srv, &serverWg)

	clientWg.Add(1)
	go client(cli, &clientWg)

	serverWg.Wait()
	clientWg.Wait()
}

func server(conn *udp.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	expectedMessages := 5

	for range expectedMessages {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		fmt.Println("<server>")
		fmt.Printf("RECEIVED [%s:%d]: %s\n", addr.IP.String(), addr.Port, string(buf[:n]))
		fmt.Printf("</server>\n\n")

		resp := "ACK: " + string(buf[:n])

		_, err = conn.WriteTo([]byte(resp), addr)
		if err != nil {
			panic(err)
		}
	}
}

func client(conn *udp.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	remoteAddr := &udp.Address{
		IP:   &net.IPAddr{IP: net.IP{127, 0, 0, 1}},
		Port: 8080,
	}

	numMessages := 5

	for i := range numMessages {
		message := fmt.Sprintf("message%d", i)

		_, err := conn.WriteTo([]byte(message), remoteAddr)
		if err != nil {
			panic(err)
		}

		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		fmt.Println("<client>")
		fmt.Printf("RECEIVED [%s:%d]: %s\n", addr.IP.String(), addr.Port, string(buf[:n]))
		fmt.Printf("</client>\n\n")
	}
}
