package example

import (
	"encoding/binary"
	"fmt"

	"github.com/darkinno-tech/hj1239-go-sdk/model"
	"github.com/darkinno-tech/hj1239-go-sdk/packet"
)

func now() model.GB1239Time {
	return model.GB1239Time{
		Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 0,
	}
}

// ExampleEncodePlatformPacket demonstrates building and encoding a platform login packet.
func ExampleEncodePlatformPacket() {
	loginData := &model.PlatformLoginCode{
		AccessTime:  now(),
		SerialNum:   1,
		Username:    "admin",
		Password:    "password123",
		EncryptMode: 0x01,
	}

	loginBytes, err := loginData.Encode()
	if err != nil {
		panic(err)
	}

	p, err := packet.BuildPlatformPacket(
		0x07,
		0xFE,
		"VIN12345678901234",
		0x01,
		loginBytes,
	)
	if err != nil {
		panic(err)
	}

	wireData, err := packet.EncodePlatformPacket(p)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Encoded %d bytes\n", len(wireData))
}

// ExampleDecodeVehicleLogin demonstrates decoding raw bytes into a vehicle login message.
func ExampleDecodeVehicleLogin() {
	raw := make([]byte, 36)
	raw[0] = 24
	raw[1] = 3
	raw[2] = 15
	raw[3] = 14
	raw[4] = 30
	raw[5] = 45
	binary.BigEndian.PutUint16(raw[6:8], 1)
	raw[28] = 1
	raw[29] = 0x01
	raw[30] = 24
	raw[31] = 3
	raw[32] = 15
	raw[33] = 14
	raw[34] = 30
	raw[35] = 46

	var v model.VehicleLoginCode
	if err := v.Decode(raw); err != nil {
		panic(err)
	}

	fmt.Printf("SerialNum: %d\n", v.SerialNum)
	fmt.Printf("Time: 20%02d-%02d-%02d %02d:%02d:%02d\n",
		v.Time.Year, v.Time.Month, v.Time.Day,
		v.Time.Hour, v.Time.Minute, v.Time.Second)
}

// ExampleEngineDataDPFSCR demonstrates encoding and decoding engine data.
func ExampleEngineDataDPFSCR() {
	e := &model.EngineDataDPFSCR{
		Speed:             model.ScaledUint16{Raw: 0x2000, Valid: true},
		AtmPressure:       model.ScaledUint8{Raw: 100, Valid: true},
		EngineSpeed:       model.ScaledUint16{Raw: 0x1000, Valid: true},
		FuelFlow:          model.ScaledUint16{Raw: 0x0100, Valid: true},
		SCRUpstreamNOx:    model.ScaledInt16{Raw: 0x0500, Valid: true},
		SCRDownstreamNOx:  model.ScaledInt16{Raw: 0x0200, Valid: true},
		SCRUpstreamTemp:   model.ScaledInt16{Raw: 0x0400, Valid: true},
		SCRDownstreamTemp: model.ScaledInt16{Raw: 0x0500, Valid: true},
		Longitude:         model.ScaledUint32{Raw: 116000000, Valid: true},
		Latitude:          model.ScaledUint32{Raw: 40000000, Valid: true},
		Odometer:          model.ScaledUint32{Raw: 1234560, Valid: true},
		PositionStatus:    0x01,
	}

	data, _ := e.Encode()
	fmt.Printf("Engine data size: %d bytes (expected 37)\n", len(data))

	var decoded model.EngineDataDPFSCR
	_ = decoded.Decode(data)

	fmt.Printf("Speed raw: 0x%04x\n", decoded.Speed.Raw)
	fmt.Printf("Engine speed raw: 0x%04x\n", decoded.EngineSpeed.Raw)
}

// ExamplePacketRoundTrip demonstrates full encode→decode round trip.
func ExamplePacketRoundTrip() {
	loginData := &model.PlatformLoginCode{
		AccessTime:  now(),
		SerialNum:   1,
		Username:    "admin",
		Password:    "pass",
		EncryptMode: 0x01,
	}
	raw, _ := loginData.Encode()

	p, _ := packet.BuildPlatformPacket(0x07, 0xFE, "VIN12345678901234", 0x01, raw)
	wireData, _ := packet.EncodePlatformPacket(p)

	parsed, err := packet.ParsePlatformPacket(wireData)
	if err != nil {
		panic(err)
	}

	if !packet.VerifyChecksum(wireData) {
		panic("checksum failed")
	}

	decoded, err := packet.DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
	if err != nil {
		panic(err)
	}

	result := decoded.(*model.PlatformLoginCode)
	fmt.Printf("Platform login: user=%s, serial=%d\n", result.Username, result.SerialNum)
}
