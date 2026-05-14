package codec

import (
	"testing"
)

func TestReadWriteByte(t *testing.T) {
	buf := make([]byte, 1)
	WriteByte(buf, 0, 0xAB)

	val, err := ReadByte(buf, 0)
	if err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if val != 0xAB {
		t.Errorf("expected 0xAB, got 0x%02x", val)
	}

	_, err = ReadByte(buf, 100)
	if err == nil {
		t.Error("expected error for out of range")
	}
}

func TestReadWriteUint16(t *testing.T) {
	buf := make([]byte, 2)
	WriteUint16(buf, 0, 0x1234)

	val, err := ReadUint16(buf, 0)
	if err != nil {
		t.Fatalf("ReadUint16: %v", err)
	}
	if val != 0x1234 {
		t.Errorf("expected 0x1234, got 0x%04x", val)
	}
}

func TestReadWriteUint32(t *testing.T) {
	buf := make([]byte, 4)
	WriteUint32(buf, 0, 0x12345678)

	val, err := ReadUint32(buf, 0)
	if err != nil {
		t.Fatalf("ReadUint32: %v", err)
	}
	if val != 0x12345678 {
		t.Errorf("expected 0x12345678, got 0x%08x", val)
	}
}

func TestReadWriteString(t *testing.T) {
	buf := make([]byte, 10)
	WriteString(buf, 0, 10, "HELLO")

	s, err := ReadString(buf, 0, 10)
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if s != "HELLO" {
		t.Errorf("expected HELLO, got %s", s)
	}
}

func TestBCC(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	checksum := BCC(data)

	// 0x01 ^ 0x02 = 0x03, 0x03 ^ 0x03 = 0x00
	if checksum != 0x00 {
		t.Errorf("expected 0x00, got 0x%02x", checksum)
	}

	// Different data
	data2 := []byte{0x01, 0x02, 0x04}
	checksum2 := BCC(data2)
	// 0x01 ^ 0x02 = 0x03, 0x03 ^ 0x04 = 0x07
	if checksum2 != 0x07 {
		t.Errorf("expected 0x07, got 0x%02x", checksum2)
	}
}

func TestEncodeDecodeScaled(t *testing.T) {
	// EncodeScaledUint16 tests
	encoded := EncodeScaledUint16(100.0, 0.05, 0)
	if encoded != 2000 {
		t.Errorf("EncodeScaledUint16: expected 2000, got %d", encoded)
	}

	decoded := DecodeScaledUint16(2000, 0.05, 0)
	if decoded != 100.0 {
		t.Errorf("DecodeScaledUint16: expected 100.0, got %f", decoded)
	}

	// Test with offset
	encoded2 := EncodeScaledInt16(25.0, 0.03125, -273)
	// (25 - (-273)) / 0.03125 = 298 / 0.03125 = 9536
	eVal := uint16(int16((25.0 - (-273.0)) / 0.03125))
	if encoded2 != eVal {
		t.Errorf("EncodeScaledInt16: expected %d, got %d", eVal, encoded2)
	}

	decoded2 := DecodeScaledInt16(eVal, 0.03125, -273)
	if decoded2 != 25.0 {
		t.Errorf("DecodeScaledInt16: expected 25.0, got %f", decoded2)
	}
}

func TestReadOutOfRange(t *testing.T) {
	buf := make([]byte, 2)

	_, err := ReadUint32(buf, 0)
	if err == nil {
		t.Error("expected error for out of range uint32")
	}

	_, err = ReadBytes(buf, 0, 100)
	if err == nil {
		t.Error("expected error for out of range bytes")
	}
}
