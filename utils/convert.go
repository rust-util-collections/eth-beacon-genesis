package utils

import "encoding/binary"

func UintToBytes(data any) []byte {
	var res []byte

	if d64, ok := data.(uint64); ok {
		res = make([]byte, 8)
		binary.LittleEndian.PutUint64(res, d64)
	} else if d32, ok := data.(uint32); ok {
		res = make([]byte, 4)
		binary.LittleEndian.PutUint32(res, d32)
	} else if d16, ok := data.(uint16); ok {
		res = make([]byte, 2)
		binary.LittleEndian.PutUint16(res, d16)
	} else if d8, ok := data.(uint8); ok {
		res = []byte{d8}
	}

	return res
}

func BytesToUint(data []byte) uint64 {
	switch len(data) {
	case 1:
		return uint64(data[0])
	case 2:
		return uint64(binary.LittleEndian.Uint16(data))
	case 4:
		return uint64(binary.LittleEndian.Uint32(data))
	case 8:
		return binary.LittleEndian.Uint64(data)
	default:
		return 0
	}
}
