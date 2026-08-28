package packet

import (
	"testing"

	"github.com/darkinno-tech/hj1239-go-sdk/model"
)

func TestEncodeDecodePlatformPacket(t *testing.T) {
	p, err := BuildPlatformPacket(
		0x01,       // 车辆登入
		0xFE,       // 请求包
		"TESTVIN1234567890",
		0x01,       // 不加密
		[]byte{0x01, 0x02, 0x03},
	)
	if err != nil {
		t.Fatalf("BuildPlatformPacket: %v", err)
	}

	data, err := EncodePlatformPacket(p)
	if err != nil {
		t.Fatalf("EncodePlatformPacket: %v", err)
	}

	// Verify structure
	if data[0] != 0x7e || data[1] != 0x7e {
		t.Errorf("invalid start marker: %02x %02x", data[0], data[1])
	}
	if data[2] != 0x01 {
		t.Errorf("invalid command flag: %02x", data[2])
	}
	if data[3] != 0xFE {
		t.Errorf("invalid response flag: %02x", data[3])
	}
	if data[21] != 0x01 {
		t.Errorf("invalid encrypt mode: %02x", data[21])
	}

	// Parse back
	parsed, err := ParsePlatformPacket(data)
	if err != nil {
		t.Fatalf("ParsePlatformPacket: %v", err)
	}

	if parsed.CommandFlag != 0x01 {
		t.Errorf("parsed command flag: %02x", parsed.CommandFlag)
	}
	if parsed.VIN != "TESTVIN1234567890" {
		t.Errorf("parsed VIN: %s", parsed.VIN)
	}
	if parsed.DataLength != 3 {
		t.Errorf("parsed data length: %d", parsed.DataLength)
	}
	if len(parsed.DataUnit) != 3 {
		t.Errorf("parsed data unit len: %d", len(parsed.DataUnit))
	}

	// Verify checksum
	if !VerifyChecksum(data) {
		t.Error("checksum verification failed")
	}
}

func TestVerifyChecksum(t *testing.T) {
	p, _ := BuildPlatformPacket(0x01, 0xFE, "VIN12345678901234", 0x01, []byte{0xAA, 0xBB})
	data, _ := EncodePlatformPacket(p)

	if !VerifyChecksum(data) {
		t.Error("checksum should be valid")
	}

	// Corrupt the data
	data[10] ^= 0xFF
	if VerifyChecksum(data) {
		t.Error("checksum should be invalid after corruption")
	}
}

func TestParsePlatformPacketInvalidMarker(t *testing.T) {
	data := make([]byte, 30)
	data[0] = 0x00
	data[1] = 0x00

	_, err := ParsePlatformPacket(data)
	if err == nil {
		t.Error("expected error for invalid start marker")
	}
}

func TestEncodePlatformPacketLongVIN(t *testing.T) {
	_, err := BuildPlatformPacket(0x01, 0xFE, "TOOLONGVIN1234567890XXXXXXXX", 0x01, nil)
	if err == nil {
		t.Error("expected error for long VIN")
	}
}

func TestDecodeDataUnitVehicleLogin(t *testing.T) {
	loginData := &model.VehicleLoginCode{
		Time:      model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 45},
		SerialNum: 1,
		ICCID:     "TESTICCID123456789",
		AuthData:  []byte{0x01},
		LoginTime: model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 46},
	}
	raw, _ := loginData.Encode()

	decoded, err := DecodeDataUnit(0x01, raw)
	if err != nil {
		t.Fatalf("DecodeDataUnit: %v", err)
	}

	v, ok := decoded.(*model.VehicleLoginCode)
	if !ok {
		t.Fatalf("expected *model.VehicleLoginCode, got %T", decoded)
	}
	if v.SerialNum != 1 {
		t.Errorf("SerialNum: expected 1, got %d", v.SerialNum)
	}
}

func TestDecodeDataUnitVehicleLogout(t *testing.T) {
	logoutData := &model.VehicleLogoutCode{
		Time:      model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 45},
		SerialNum: 1,
	}
	raw, _ := logoutData.Encode()

	decoded, err := DecodeDataUnit(0x04, raw)
	if err != nil {
		t.Fatalf("DecodeDataUnit: %v", err)
	}

	v, ok := decoded.(*model.VehicleLogoutCode)
	if !ok {
		t.Fatalf("expected *model.VehicleLogoutCode, got %T", decoded)
	}
	if v.SerialNum != 1 {
		t.Errorf("SerialNum: expected 1, got %d", v.SerialNum)
	}
}

func TestDecodeDataUnitUnknown(t *testing.T) {
	_, err := DecodeDataUnit(0xFF, []byte{0x00})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestDecodeDataUnitPlatformLogin(t *testing.T) {
	loginData := &model.PlatformLoginCode{
		AccessTime:  model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 45},
		SerialNum:   1,
		Username:    "admin",
		Password:    "password123",
		EncryptMode: 0x01,
	}
	raw, _ := loginData.Encode()

	decoded, err := DecodeDataUnit(0x07, raw)
	if err != nil {
		t.Fatalf("DecodeDataUnit: %v", err)
	}

	p, ok := decoded.(*model.PlatformLoginCode)
	if !ok {
		t.Fatalf("expected *model.PlatformLoginCode, got %T", decoded)
	}
	if p.Username != "admin" {
		t.Errorf("Username: expected admin, got %s", p.Username)
	}
}
