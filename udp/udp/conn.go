package udp

import (
	"encoding/binary"
	"errors"
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

type Conn struct {
	sync.Mutex

	ipListener *net.IPConn
	listenPort int
	listenIP   *net.IPAddr

	packetChan chan *message
	errChan    chan error

	closeChan chan struct{}
}

func newConn(ipListener *net.IPConn, address *Address) *Conn {
	result := &Conn{
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

func (l *Conn) Close() error {
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

func (l *Conn) readLoop() {
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

func (l *Conn) processData(data []byte, n int, sourceAddr *net.IPAddr, err error) (*message, bool) {
	if err != nil {
		return &message{Err: err}, true
	}

	data = data[:n]

	if len(data) < 8 {
		return nil, false
	}

	sourcePort := binary.BigEndian.Uint16(data[0:2])
	destPort := binary.BigEndian.Uint16(data[2:4])
	checksum := binary.BigEndian.Uint16(data[6:8])

	calculatedSum := calculateReceiveChecksum(data, sourceAddr.IP.To4(), l.listenIP.IP.To4())
	calculatedPseudoSum := calculatePseudoheaderSum(data, sourceAddr.IP.To4(), l.listenIP.IP.To4())

	if calculatedSum+checksum != 0xffff && calculatedPseudoSum != checksum {
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

func (l *Conn) ReadFrom(b []byte) (int, *Address, error) {
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

func (l *Conn) WriteTo(b []byte, address *Address) (int, error) {
	packet := addHeader(&headerParams{
		Data: b,
		Source: &Address{
			IP:   l.listenIP,
			Port: l.listenPort,
		},
		Dest: address,
	})

	return l.ipListener.WriteToIP(packet, address.IP)
}

func Listen(network string, address *Address) (*Conn, error) {
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

	return newConn(ipConn, address), nil
}
