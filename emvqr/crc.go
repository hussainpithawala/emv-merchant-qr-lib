package emvqr

import "fmt"

// crc16CCITT computes the CRC-16/CCITT-FALSE checksum used by EMV QR codes.
//
// Polynomial : 0x1021
// Initial value: 0xFFFF
// Input/Output reflection: none
// XOR out: 0x0000
func crc16CCITT(data []byte) uint16 {
	const poly = uint16(0x1021)
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// crcString encodes a uint16 as a 4-character upper-case hex string.
func crcString(v uint16) string {
	return fmt.Sprintf("%04X", v)
}
