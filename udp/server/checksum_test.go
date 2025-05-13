package server

import (
	"net"
	"testing"
)

func TestCalculateReceiveChecksum_InputPacketWithNoPadding_GetCorrectCheckSum(t *testing.T) {
	sourceIP := net.IP{152, 1, 51, 27}
	destIP := net.IP{152, 14, 94, 75}
	data := []byte{0xa0, 0x8f, 0x26, 0x94, 0x00, 0x0a, 0x00, 0x00, 0x62, 0x62}

	result := calculateReceiveChecksum(data, sourceIP, destIP)

	if result != 0xeb21 {
		t.Fatalf("expected %x, got %x", 0xeb21, result)
	}
}

func TestCalculateSendChecksum_InputPacketWithNoPadding_GetCorrectCheckSum(t *testing.T) {
	sourceIP := net.IP{152, 1, 51, 27}
	destIP := net.IP{152, 14, 94, 75}
	data := []byte{0xa0, 0x8f, 0x26, 0x94, 0x00, 0x0a, 0x00, 0x00, 0x62, 0x62}

	result := calculateSendChecksum(data, sourceIP, destIP)

	if result != 0x14de {
		t.Fatalf("expected %x, got %x", 0x14de, result)
	}
}
