package codec

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ByteOrder defines encoding byte order. GB1239 uses Big-Endian.
var ByteOrder = binary.BigEndian

// ReadByte 读取单字节
func ReadByte(data []byte, offset int) (byte, error) {
	if offset >= len(data) {
		return 0, fmt.Errorf("read byte at %d: out of range (len=%d)", offset, len(data))
	}
	return data[offset], nil
}

// WriteByte 写入单字节
func WriteByte(buf []byte, offset int, value byte) {
	buf[offset] = value
}

// ReadUint16 读取大端 uint16
func ReadUint16(data []byte, offset int) (uint16, error) {
	if offset+2 > len(data) {
		return 0, fmt.Errorf("read uint16 at %d: out of range (len=%d)", offset, len(data))
	}
	return ByteOrder.Uint16(data[offset:]), nil
}

// WriteUint16 写入大端 uint16
func WriteUint16(buf []byte, offset int, value uint16) {
	ByteOrder.PutUint16(buf[offset:offset+2], value)
}

// ReadUint32 读取大端 uint32
func ReadUint32(data []byte, offset int) (uint32, error) {
	if offset+4 > len(data) {
		return 0, fmt.Errorf("read uint32 at %d: out of range (len=%d)", offset, len(data))
	}
	return ByteOrder.Uint32(data[offset:]), nil
}

// WriteUint32 写入大端 uint32
func WriteUint32(buf []byte, offset int, value uint32) {
	ByteOrder.PutUint32(buf[offset:offset+4], value)
}

// ReadBytes 读取指定长度的字节
func ReadBytes(data []byte, offset int, length int) ([]byte, error) {
	if offset+length > len(data) {
		return nil, fmt.Errorf("read %d bytes at %d: out of range (len=%d)", length, offset, len(data))
	}
	b := make([]byte, length)
	copy(b, data[offset:offset+length])
	return b, nil
}

// WriteBytes 写入字节数组
func WriteBytes(buf []byte, offset int, data []byte) {
	copy(buf[offset:], data)
}

// ReadString 读取定长字符串（去除尾部空字节）
func ReadString(data []byte, offset int, length int) (string, error) {
	b, err := ReadBytes(data, offset, length)
	if err != nil {
		return "", err
	}
	// Trim trailing null bytes
	for i, c := range b {
		if c == 0 {
			return string(b[:i]), nil
		}
	}
	return string(b), nil
}

// WriteString 写入定长字符串（不足用0填充）
func WriteString(buf []byte, offset int, length int, s string) {
	written := copy(buf[offset:offset+length], s)
	for i := written; i < length; i++ {
		buf[offset+i] = 0
	}
}

// EncodeScaledUint8 编码带缩放的无符号 uint8
func EncodeScaledUint8(value float64, scale float64, offsetVal float64) uint8 {
	if math.IsNaN(value) {
		return 0xFF
	}
	if value < offsetVal {
		return 0
	}
	raw := uint8((value - offsetVal) / scale)
	return raw
}

func DecodeScaledUint8(raw uint8, scale float64, offsetVal float64) float64 {
	if raw == 0xFF {
		return math.NaN()
	}
	return float64(raw)*scale + offsetVal
}

func EncodeScaledInt8(value float64, scale float64, offsetVal float64) uint8 {
	if math.IsNaN(value) {
		return 0xFF
	}
	v := int8((value - offsetVal) / scale)
	raw := uint8(v)
	if raw == 0xFF {
		return 0xFE // avoid invalid marker collision
	}
	return raw
}

func DecodeScaledInt8(raw uint8, scale float64, offsetVal float64) float64 {
	if raw == 0xFF {
		return math.NaN()
	}
	return float64(int8(raw))*scale + offsetVal
}

func EncodeScaledUint16(value float64, scale float64, offsetVal float64) uint16 {
	if math.IsNaN(value) {
		return 0xFFFF
	}
	if value < offsetVal {
		return 0
	}
	raw := uint16((value - offsetVal) / scale)
	return raw
}

func DecodeScaledUint16(raw uint16, scale float64, offsetVal float64) float64 {
	if raw == 0xFFFF {
		return math.NaN()
	}
	return float64(raw)*scale + offsetVal
}

func EncodeScaledInt16(value float64, scale float64, offsetVal float64) uint16 {
	if math.IsNaN(value) {
		return 0xFFFF
	}
	v := int16((value - offsetVal) / scale)
	raw := uint16(v)
	if raw == 0xFFFF {
		return 0xFFFE // avoid invalid marker collision
	}
	return raw
}

func DecodeScaledInt16(raw uint16, scale float64, offsetVal float64) float64 {
	if raw == 0xFFFF {
		return math.NaN()
	}
	return float64(int16(raw))*scale + offsetVal
}

func EncodeScaledUint32(value float64, scale float64, offsetVal float64) uint32 {
	if math.IsNaN(value) {
		return 0xFFFFFFFF
	}
	if value < offsetVal {
		return 0
	}
	raw := uint32((value - offsetVal) / scale)
	return raw
}

func DecodeScaledUint32(raw uint32, scale float64, offsetVal float64) float64 {
	if raw == 0xFFFFFFFF {
		return math.NaN()
	}
	return float64(raw)*scale + offsetVal
}

// BCC computes BCC (Block Check Character) - XOR checksum
func BCC(data []byte) byte {
	var checksum byte
	for _, b := range data {
		checksum ^= b
	}
	return checksum
}
