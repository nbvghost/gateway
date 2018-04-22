package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"strings"
)

type uuID []byte

func (uuid uuID) String() string {
	if len(uuid) != 16 {
		return ""
	}
	var buf [36]byte
	encodeHex(buf[:], uuid)
	uu := string(buf[:])
	return strings.Replace(uu, "-", "", -1)
}
func encodeHex(dst []byte, uuid uuID) {
	hex.Encode(dst[:], uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}
func UUID() uuID {
	uuid := make([]byte, 16)
	randomBits([]byte(uuid))
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid
}
func randomBits(b []byte) {
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err.Error()) // rand should never fail
	}

}
