package utils

import "encoding/binary"

func MarshalUint32(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

func UnmarshalUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func MarshalUint64(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func UnmarshalUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func MarshalString(str string) []byte {
	if str == "" {
		return []byte{0, 0, 0, 0}
	}
	return append(MarshalUint32(uint32(len(str))), []byte(str)...)
}
