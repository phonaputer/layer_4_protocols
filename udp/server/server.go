package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
)

const (
	maxUDPPacketSize = 65535
)

type Address struct {
	IP   *net.IPAddr
	Port int
}

type UDPListener struct {
	sync.Mutex

	ipListener *net.IPConn
	listenPort int
	listenIP   *net.IPAddr

	packetChan chan *message
	errChan    chan error

	closeChan chan struct{}
}

func newUDPListener(ipListener *net.IPConn, address *Address) *UDPListener {
	result := &UDPListener{
		ipListener: ipListener,
		listenPort: address.Port,
		listenIP:   address.IP,
		packetChan: make(chan *message),
		errChan:    make(chan error),
		closeChan:  make(chan struct{}),
	}

	go result.readLoop()

	return result
}

func (l *UDPListener) Close() error {
	l.Lock()
	defer l.Unlock()

	close(l.closeChan)

	return nil
}

type message struct {
	Data    []byte
	Address *Address
	Err     error
}

func (l *UDPListener) readLoop() {
	for {
		buffer := make([]byte, maxUDPPacketSize)
		n, addr, err := l.ipListener.ReadFromIP(buffer)

		msg, ok := l.processData(buffer, n, addr, err)
		if ok {
			select {
			case <-l.closeChan:
				return
			case l.packetChan <- msg:
			}
		}
	}
}

func (l *UDPListener) processData(data []byte, n int, sourceAddr *net.IPAddr, err error) (*message, bool) {
	if err != nil {
		return &message{Err: err}, true
	}

	data = data[:n]

	fmt.Println(string(data))

	if len(data) < 8 {
		return nil, false
	}

	sourcePort := binary.BigEndian.Uint16(data[0:2])
	destPort := binary.BigEndian.Uint16(data[2:4])
	length := binary.BigEndian.Uint16(data[4:6])
	checksum := binary.BigEndian.Uint16(data[6:8])

	calcSum := calculateSendChecksum(data, sourceAddr.IP.To4(), l.listenIP.IP.To4())

	fmt.Println("source:", sourceAddr.String(), ", dest:", l.listenIP.String())

	fmt.Println("sourcePort:", sourcePort, ", destPort:", destPort, ", length:", length)
	fmt.Printf("checksum: %x\n", checksum)
	fmt.Printf("calcsum: %x\n", calcSum)

	fmt.Printf("sum: %x\n", checksum&calcSum)

	if calcSum != checksum {
		return nil, false
	}

	if int(destPort) != l.listenPort {
		return nil, false
	}

	return &message{
		Data: data[8:],
		Address: &Address{
			IP:   sourceAddr,
			Port: int(sourcePort),
		},
	}, true
}

func (l *UDPListener) ReadFrom(b []byte) (int, *Address, error) {
	var msg *message
	select {
	case chanMsg := <-l.packetChan:
		msg = chanMsg
	case <-l.closeChan:
		return 0, nil, errors.New("listener closed")
	}

	if msg.Err != nil {
		return 0, nil, msg.Err
	}

	return copy(b, msg.Data), msg.Address, nil
}

func ListenUDP(network string, address *Address) (*UDPListener, error) {
	var ipNetwork string
	switch network {
	case "udp":
		ipNetwork = "ip"
	case "udp4":
		ipNetwork = "ip4"
	case "udp6":
		ipNetwork = "ip6"
	default:
		return nil, errors.New("unknown network")
	}

	ipConn, err := net.ListenIP(
		ipNetwork+":udp",
		address.IP,
	)
	if err != nil {
		return nil, err
	}

	return newUDPListener(ipConn, address), nil
}
