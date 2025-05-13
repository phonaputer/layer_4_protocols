package server

import (
	"encoding/binary"
	"net"
)

// based on https://cs.newpaltz.edu/~easwarac/CCN/Week5/Checksum.pdf
// and http://profesores.elo.utfsm.cl/~agv/elo322/UDP_Checksum_HowTo.html
// and https://gist.github.com/david-hoze/0c7021434796997a4ca42d7731a7073a
// Notes about partial offloading here: https://wiki.wireshark.org/CaptureSetup/Offloading
func calculateReceiveChecksum(data []byte, sourceAddr net.IP, destAddr net.IP) uint16 {
	var sum uint32
	var dataToAdd []byte
	dataToAdd = append(dataToAdd, data[:6]...)
	dataToAdd = append(dataToAdd, data[8:]...)

	if len(dataToAdd)%2 != 0 {
		dataToAdd = append(dataToAdd, 0x00)
	}

	for itr := 0; itr < len(dataToAdd); itr += 2 {
		sum += uint32(binary.BigEndian.Uint16(dataToAdd[itr : itr+2]))
	}

	sum = sum + uint32(calculatePseudoheaderSum(data, sourceAddr, destAddr))

	checkSum := sum & 0xffff
	checkSum += sum >> 16 & 0xffff

	result := uint16(checkSum)

	if result == 0x0000 {
		return 0xffff
	}

	return result
}

func calculatePseudoheaderSum(data []byte, sourceAddr net.IP, destAddr net.IP) uint16 {
	length := binary.BigEndian.Uint16(data[4:6])
	source := binary.BigEndian.Uint32(sourceAddr)
	dest := binary.BigEndian.Uint32(destAddr)

	// Add pseudo header
	var sum uint32
	sum += uint32(length)
	sum += source & 0xffff
	sum += source >> 16 & 0xffff
	sum += dest & 0xffff
	sum += dest >> 16 & 0xffff
	sum += 0x0011

	checkSum := sum & 0xffff
	checkSum += sum >> 16 & 0xffff

	return uint16(checkSum)
}

func calculateSendChecksum(data []byte, sourceAddr net.IP, destAddr net.IP) uint16 {
	return ^calculateReceiveChecksum(data, sourceAddr, destAddr)
}
