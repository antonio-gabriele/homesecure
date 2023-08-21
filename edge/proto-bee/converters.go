package main

import (
	"fmt"
	"strconv"
	"strings"
)

func HexToUInt8(src string) uint8 {
	hex, _ := strings.CutPrefix(src, "0x")
	endpoint, _ := strconv.ParseUint(hex, 16, 8)
	return uint8(endpoint)
}

func HexToUInt16(src string) uint16 {
	hex, _ := strings.CutPrefix(src, "0x")
	endpoint, _ := strconv.ParseUint(hex, 16, 16)
	return uint16(endpoint)
}

func HexToUInt64(src string) uint64 {
	hex, _ := strings.CutPrefix(src, "0x")
	endpoint, _ := strconv.ParseUint(hex, 16, 64)
	return endpoint
}

func UInt8ToHex(src uint8) string {
	return fmt.Sprintf("%02x", src)
}

func UInt16ToHex(src uint16) string {
	return fmt.Sprintf("%04x", src)
}

/*
func UInt64ToHex(src uint64) string {
	return fmt.Sprintf("%016x", src)
}
*/
