package udp

import (
	"encoding/binary"
)

const (
	headerSize = 8
)

type headerParams struct {
	Data   []byte
	Source *Address
	Dest   *Address
}

func addHeader(params *headerParams) []byte {
	packet := make([]byte, headerSize)

	binary.BigEndian.PutUint16(packet[:2], uint16(params.Source.Port))
	binary.BigEndian.PutUint16(packet[2:4], uint16(params.Dest.Port))

	length := len(params.Data) + headerSize
	binary.BigEndian.PutUint16(packet[4:6], uint16(length))

	packet = append(packet, params.Data...)

	checkSum := calculateSendChecksum(packet, params.Source.IP.IP.To4(), params.Dest.IP.IP.To4())
	binary.BigEndian.PutUint16(packet[6:8], checkSum)

	return packet
}
