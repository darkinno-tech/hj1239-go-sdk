package model

import (
	"testing"
	"time"
)

func TestGB1239Time(t *testing.T) {
	tm := time.Date(2024, 3, 15, 14, 30, 45, 0, time.FixedZone("CST", 8*3600))
	gt := NewGB1239Time(tm)

	if gt.Year != 24 {
		t.Errorf("expected year 24, got %d", gt.Year)
	}
	if gt.Month != 3 {
		t.Errorf("expected month 3, got %d", gt.Month)
	}
	if gt.Day != 15 {
		t.Errorf("expected day 15, got %d", gt.Day)
	}
	if gt.Hour != 14 {
		t.Errorf("expected hour 14, got %d", gt.Hour)
	}
	if gt.Minute != 30 {
		t.Errorf("expected minute 30, got %d", gt.Minute)
	}
	if gt.Second != 45 {
		t.Errorf("expected second 45, got %d", gt.Second)
	}

	// Test bytes round-trip
	b := gt.Bytes()
	if len(b) != 6 {
		t.Errorf("expected 6 bytes, got %d", len(b))
	}

	gt2, err := ParseGB1239Time(b)
	if err != nil {
		t.Fatalf("ParseGB1239Time: %v", err)
	}
	if gt2.Year != 24 || gt2.Month != 3 {
		t.Errorf("round-trip failed")
	}

	// Test time.Time round-trip
	tm2 := gt2.Time()
	if tm2.Year() != 2024 || tm2.Month() != 3 || tm2.Day() != 15 {
		t.Errorf("time round-trip failed: %v", tm2)
	}
}

func TestScaledTypes(t *testing.T) {
	// Test ScaledUint8
	s8 := ScaledUint8{Valid: true, Raw: 100}
	buf := make([]byte, 1)
	s8.PutUint8(buf)
	if buf[0] != 100 {
		t.Errorf("ScaledUint8 put: expected 100, got %d", buf[0])
	}

	var s8d ScaledUint8
	s8d.DecodeUint8(buf)
	if s8d.Raw != 100 {
		t.Errorf("ScaledUint8 decode: expected 100, got %d", s8d.Raw)
	}

	// Test invalid value
	invalid := ScaledUint8{Valid: false}
	invalid.PutUint8(buf)
	if buf[0] != 0xFF {
		t.Errorf("ScaledUint8 invalid put: expected 0xFF, got %02x", buf[0])
	}
}

func TestEngineDataDPFSCREncodeDecode(t *testing.T) {
	e := &EngineDataDPFSCR{
		Speed:             ScaledUint16{Raw: 0x1234, Valid: true},
		AtmPressure:       ScaledUint8{Raw: 0x64, Valid: true},
		EngineTorque:      ScaledInt8{Raw: 0x50, Valid: true},
		FrictionTorque:    ScaledInt8{Raw: 0x30, Valid: true},
		EngineSpeed:       ScaledUint16{Raw: 0x2000, Valid: true},
		FuelFlow:          ScaledUint16{Raw: 0x0100, Valid: true},
		SCRUpstreamNOx:    ScaledInt16{Raw: 0x0500, Valid: true},
		SCRDownstreamNOx:  ScaledInt16{Raw: 0x0200, Valid: true},
		ReagentRate:       ScaledUint8{Raw: 0x80, Valid: true},
		AirFlow:           ScaledUint16{Raw: 0x0300, Valid: true},
		SCRUpstreamTemp:   ScaledInt16{Raw: 0x0400, Valid: true},
		SCRDownstreamTemp: ScaledInt16{Raw: 0x0500, Valid: true},
		DPFDiffPressure:   ScaledUint16{Raw: 0x0050, Valid: true},
		CoolantTemp:       ScaledInt8{Raw: 0x60, Valid: true},
		UreaLevel:         ScaledUint8{Raw: 0x90, Valid: true},
		PositionStatus:    0x01,
		Longitude:         ScaledUint32{Raw: 116000000, Valid: true}, // 116.0 deg
		Latitude:          ScaledUint32{Raw: 40000000, Valid: true},  // 40.0 deg
		Odometer:          ScaledUint32{Raw: 1234560, Valid: true},   // 123456.0 km
	}

	data, err := e.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(data) != EngineDataDPFSCRSize {
		t.Errorf("expected %d bytes, got %d", EngineDataDPFSCRSize, len(data))
	}

	var decoded EngineDataDPFSCR
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.Speed.Raw != 0x1234 {
		t.Errorf("Speed: expected 0x1234, got 0x%04x", decoded.Speed.Raw)
	}
	if decoded.Odometer.Raw != 1234560 {
		t.Errorf("Odometer: expected 1234560, got %d", decoded.Odometer.Raw)
	}
}

func TestVehicleLoginCode(t *testing.T) {
	v := &VehicleLoginCode{
		Time: GB1239Time{
			Year: 24, Month: 3, Day: 15,
			Hour: 14, Minute: 30, Second: 45,
		},
		SerialNum: 12345,
		ICCID:     "12345678901234567890",
		AuthData:  []byte{0x01, 0x02, 0x03},
		LoginTime: GB1239Time{
			Year: 24, Month: 3, Day: 15,
			Hour: 14, Minute: 30, Second: 46,
		},
	}

	data, err := v.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	var decoded VehicleLoginCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.SerialNum != 12345 {
		t.Errorf("SerialNum: expected 12345, got %d", decoded.SerialNum)
	}
	if decoded.ICCID != "12345678901234567890" {
		t.Errorf("ICCID: expected 12345678901234567890, got %s", decoded.ICCID)
	}
	if len(decoded.AuthData) != 3 {
		t.Errorf("AuthData len: expected 3, got %d", len(decoded.AuthData))
	}
}

func TestVehicleLogoutCode(t *testing.T) {
	v := &VehicleLogoutCode{
		Time:      GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 45},
		SerialNum: 12345,
	}

	data, err := v.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(data) != 8 {
		t.Errorf("expected 8 bytes, got %d", len(data))
	}

	var decoded VehicleLogoutCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.SerialNum != 12345 {
		t.Errorf("SerialNum: expected 67890, got %d", decoded.SerialNum)
	}
}

func TestPlatformLoginCode(t *testing.T) {
	p := &PlatformLoginCode{
		AccessTime:  GB1239Time{Year: 24, Month: 1, Day: 15, Hour: 10, Minute: 0, Second: 0},
		SerialNum:   1,
		Username:    "admin",
		Password:    "password123",
		EncryptMode: 0x01,
	}

	data, err := p.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(data) != PlatformLoginCodeSize {
		t.Errorf("expected %d bytes, got %d", PlatformLoginCodeSize, len(data))
	}

	var decoded PlatformLoginCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.SerialNum != 1 {
		t.Errorf("SerialNum: expected 1, got %d", decoded.SerialNum)
	}
	if decoded.Username != "admin" {
		t.Errorf("Username: expected admin, got %s", decoded.Username)
	}
	if decoded.EncryptMode != 0x01 {
		t.Errorf("EncryptMode: expected 0x01, got 0x%02x", decoded.EncryptMode)
	}
}

func TestScaleHelpers(t *testing.T) {
	// Test scale encoding/decoding
	v := ScaleUint16(0x1234, 0.05, 0)
	t.Logf("ScaledUint16 value: %f", v)

	// Test encode
	encoded := EncodeUint16(100.0, 0.05, 0)
	if encoded != 2000 {
		t.Errorf("EncodeUint16: expected 2000, got %d", encoded)
	}

	// Test invalid
	invalid := EncodeUint16(0, 1, 0)
	if invalid == 0xFFFF {
		t.Error("EncodeUint16: should not return 0xFFFF for valid value")
	}
}

func TestRealTimeInfoReport(t *testing.T) {
	engineData := &EngineDataDPFSCR{
		Speed:             ScaledUint16{Raw: 0x2000, Valid: true},
		AtmPressure:       ScaledUint8{Raw: 100, Valid: true},
		EngineTorque:      ScaledInt8{Raw: 0x50, Valid: true},
		FrictionTorque:    ScaledInt8{Raw: 0x30, Valid: true},
		EngineSpeed:       ScaledUint16{Raw: 0x2000, Valid: true},
		FuelFlow:          ScaledUint16{Raw: 0x100, Valid: true},
		SCRUpstreamNOx:    ScaledInt16{Raw: 0x500, Valid: true},
		SCRDownstreamNOx:  ScaledInt16{Raw: 0x200, Valid: true},
		ReagentRate:       ScaledUint8{Raw: 0x80, Valid: true},
		AirFlow:           ScaledUint16{Raw: 0x300, Valid: true},
		SCRUpstreamTemp:   ScaledInt16{Raw: 0x400, Valid: true},
		SCRDownstreamTemp: ScaledInt16{Raw: 0x500, Valid: true},
		DPFDiffPressure:   ScaledUint16{Raw: 0x50, Valid: true},
		CoolantTemp:       ScaledInt8{Raw: 0x60, Valid: true},
		UreaLevel:         ScaledUint8{Raw: 0x90, Valid: true},
		PositionStatus:    0x01,
		Longitude:         ScaledUint32{Raw: 116000000, Valid: true},
		Latitude:          ScaledUint32{Raw: 40000000, Valid: true},
		Odometer:          ScaledUint32{Raw: 1234560, Valid: true},
	}

	engineRaw, _ := engineData.Encode()

	block := RealTimeInfoBlock{
		InfoTypeFlag: 0x02,
		CollectTime:  GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 45},
		infoBodyRaw:  engineRaw,
	}

	report := &RealTimeInfoReport{
		DataTime:  GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 14, Minute: 30, Second: 50},
		SerialNum: 1,
		InfoBlocks: []RealTimeInfoBlock{block},
	}

	data, err := report.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	_ = data
	if len(data) < 8 {
		t.Errorf("expected at least 8 bytes, got %d", len(data))
	}

	// Decode back using the real decoder (validates multi-block boundary detection)
	var decoded RealTimeInfoReport
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if decoded.SerialNum != 1 {
		t.Errorf("SerialNum: expected 1, got %d", decoded.SerialNum)
	}
	if len(decoded.InfoBlocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(decoded.InfoBlocks))
	}
	t.Logf("Report encoded: %d bytes, decoded time: %v", len(data), decoded.DataTime)
}

func TestNewScaledUint16(t *testing.T) {
	s := NewScaledUint16(100.0, 0.05, 0)
	if !s.Valid {
		t.Error("expected valid")
	}
	if s.Raw != 2000 {
		t.Errorf("raw: expected 2000, got %d", s.Raw)
	}
	if s.Value != 100.0 {
		t.Errorf("value: expected 100.0, got %f", s.Value)
	}

	buf := make([]byte, 2)
	s.PutUint16(buf, true)
	if buf[0] != 0x07 || buf[1] != 0xD0 {
		t.Errorf("bytes: expected 07d0, got %02x%02x", buf[0], buf[1])
	}

	// Round trip with DecodeUint16Scaled
	var s2 ScaledUint16
	s2.DecodeUint16Scaled(buf, true, 0.05, 0)
	if s2.Value != 100.0 {
		t.Errorf("decoded value: expected 100.0, got %f", s2.Value)
	}
}

func TestNewScaledInt16(t *testing.T) {
	s := NewScaledInt16(25.0, 0.03125, -273)
	if !s.Valid {
		t.Error("expected valid")
	}
	t.Logf("Raw: %d, Value: %f", s.Raw, s.Value)
}

func TestPositionStatus(t *testing.T) {
	ps := ParsePositionStatus(0x05) // bit0=1(valid), bit2=1(north)
	if !ps.IsValid() {
		t.Error("expected valid position")
	}
	if !ps.IsNorthLat() {
		t.Error("expected north latitude")
	}
	if ps.IsWestLng() {
		t.Error("expected east longitude")
	}
}

func TestOBDInfoBody(t *testing.T) {
	o := &OBDInfoBody{
		MilStatus:          0x0001,
		DiagSupportStatus:  0xFFFF,
		DiagCompleteStatus: 0x0000,
		VIN:                "VIN12345678901234",
		DTCCount:           1,
		DTCCodes:           []byte{0x01, 0x02, 0x03, 0x04},
	}

	data, err := o.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(data) != 100 {
		t.Errorf("expected 100 bytes, got %d", len(data))
	}

	var decoded OBDInfoBody
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.MilStatus != 0x0001 {
		t.Errorf("MilStatus: expected 0x0001, got 0x%04x", decoded.MilStatus)
	}
	if decoded.VIN != "VIN12345678901234" {
		t.Errorf("VIN: expected VIN12345678901234, got %s", decoded.VIN)
	}
	if decoded.DTCCount != 1 {
		t.Errorf("DTCCount: expected 1, got %d", decoded.DTCCount)
	}
}

func TestParseInfoBody(t *testing.T) {
	// Engine DPF/SCR
	engine := &EngineDataDPFSCR{
		Speed:         ScaledUint16{Raw: 0x2000, Valid: true},
		EngineSpeed:   ScaledUint16{Raw: 0x1000, Valid: true},
		PositionStatus: 0x01,
		Longitude:     ScaledUint32{Raw: 116000000, Valid: true},
		Latitude:      ScaledUint32{Raw: 40000000, Valid: true},
	}
	raw, _ := engine.Encode()

	decoded, err := ParseInfoBody(0x02, raw)
	if err != nil {
		t.Fatalf("ParseInfoBody 0x02: %v", err)
	}
	e, ok := decoded.(*EngineDataDPFSCR)
	if !ok {
		t.Fatalf("expected *EngineDataDPFSCR, got %T", decoded)
	}
	if e.Speed.Raw != 0x2000 {
		t.Errorf("Speed raw: expected 0x2000, got 0x%04x", e.Speed.Raw)
	}

	// OBD
	obd := &OBDInfoBody{MilStatus: 0x0001, DTCCount: 0}
	rawObd, _ := obd.Encode()

	dec, err := ParseInfoBody(0x01, rawObd)
	if err != nil {
		t.Fatalf("ParseInfoBody 0x01: %v", err)
	}
	_, ok = dec.(*OBDInfoBody)
	if !ok {
		t.Fatalf("expected *OBDInfoBody, got %T", dec)
	}

	// Unknown type
	_, err = ParseInfoBody(0xFF, []byte{})
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestNewEngineDataDPFSCR(t *testing.T) {
	e := NewEngineDataDPFSCR(
		60.0, 1500.0, 101.3, 20.0, // speed, rpm, atm, fuel
		500.0, 200.0, 250.0, 200.0, // SCR NOx, temps
		5.0, 90.0, 80.0, // DPF, coolant, urea
		50.0, 20.0, 120.0, // torque %
		116.3912, 39.9073, 12345.6, // lon, lat, odo
		true, true, // valid, north
	)

	if !e.Speed.Valid {
		t.Error("speed should be valid")
	}
	if !e.Longitude.Valid {
		t.Error("longitude should be valid")
	}

	data, err := e.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(data) != EngineDataDPFSCRSize {
		t.Errorf("size: expected %d, got %d", EngineDataDPFSCRSize, len(data))
	}
}

func TestProtocolError(t *testing.T) {
	err := ErrShortData("EngineDataDPFSCR", 37, 20)
	if err.Code != ErrCodeShortData {
		t.Errorf("code: expected SHORT_DATA, got %s", err.Code)
	}
	t.Logf("Error: %v", err)

	err2 := ErrUnknownCmd(0xFF)
	if err2.Code != ErrCodeUnknownCmd {
		t.Errorf("code: expected UNKNOWN_CMD, got %s", err2.Code)
	}
}

func TestMILMasks(t *testing.T) {
	flags := uint16(MilBitDPF | MilBitSCRNOxAdsorber | MilBitMisfireSupport)
	if !HasMIL(flags, MilBitDPF) {
		t.Error("expected DPF MIL set")
	}
	if !HasMIL(flags, MilBitSCRNOxAdsorber) {
		t.Error("expected SCR/NOx MIL set")
	}
	if HasMIL(flags, MilBitCatalyst) {
		t.Error("catalyst MIL should not be set")
	}
}

func TestEncodeWithError(t *testing.T) {
	_, err := EncodeUint16WithError(100.0, 0.05, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = EncodeUint16WithError(-1.0, 0.05, 0)
	if err == nil {
		t.Error("expected error for negative value")
	}
}

func BenchmarkEngineDataDPFSCREncode(b *testing.B) {
	e := NewEngineDataDPFSCR(
		60.0, 1500.0, 101.3, 20.0,
		500.0, 200.0, 250.0, 200.0,
		5.0, 90.0, 80.0,
		50.0, 20.0, 120.0,
		116.3912, 39.9073, 12345.6,
		true, true,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Encode()
	}
}

func BenchmarkEngineDataDPFSCRDecode(b *testing.B) {
	e := NewEngineDataDPFSCR(
		60.0, 1500.0, 101.3, 20.0,
		500.0, 200.0, 250.0, 200.0,
		5.0, 90.0, 80.0,
		50.0, 20.0, 120.0,
		116.3912, 39.9073, 12345.6,
		true, true,
	)
	data, _ := e.Encode()
	var dec EngineDataDPFSCR
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dec.Decode(data)
	}
}

func TestHistoricalInfoCode(t *testing.T) {
	block1 := HistoricalInfoBlock{
		BeginTime: GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0},
		EndTime:   GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 9, Minute: 0, Second: 0},
		Data:      []byte{0x01, 0x02, 0x03, 0x04},
	}
	block2 := HistoricalInfoBlock{
		BeginTime: GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 9, Minute: 0, Second: 0},
		EndTime:   GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 10, Minute: 0, Second: 0},
		Data:      []byte{0x05, 0x06},
	}

	code := &HistoricalInfoCode{
		DataTime:  GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 10, Minute: 30, Second: 0},
		SerialNum: 42,
		Blocks:    []HistoricalInfoBlock{block1, block2},
	}

	data, err := code.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded HistoricalInfoCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.SerialNum != 42 {
		t.Errorf("SerialNum: expected 42, got %d", decoded.SerialNum)
	}
	if len(decoded.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(decoded.Blocks))
	}
	if !bytesEqual(decoded.Blocks[0].Data, block1.Data) {
		t.Errorf("block 1 data mismatch")
	}
	if !bytesEqual(decoded.Blocks[1].Data, block2.Data) {
		t.Errorf("block 2 data mismatch")
	}
	t.Logf("Historical info: encoded=%d bytes, decoded 2 blocks", len(data))
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAlarmInfoCode(t *testing.T) {
	code := &AlarmInfoCode{
		DataTime:   GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0},
		SerialNum:  1,
		AlarmCount: 2,
		Alarms: []Alarm{
			{
				AlarmType: AlarmTypeOverspeed,
				AlarmTime: GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 1},
				AlarmData: []byte{0x01, 0x02},
			},
			{
				AlarmType: AlarmTypeTamper,
				AlarmTime: GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 5},
				AlarmData: []byte{0x03},
			},
		},
	}

	data, err := code.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded AlarmInfoCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.SerialNum != 1 {
		t.Errorf("SerialNum: expected 1, got %d", decoded.SerialNum)
	}
	if len(decoded.Alarms) != 2 {
		t.Fatalf("expected 2 alarms, got %d", len(decoded.Alarms))
	}
	if decoded.Alarms[0].AlarmType != AlarmTypeOverspeed {
		t.Errorf("alarm 0 type: expected 0x04, got 0x%02x", decoded.Alarms[0].AlarmType)
	}
	if !bytesEqual(decoded.Alarms[0].AlarmData, []byte{0x01, 0x02}) {
		t.Errorf("alarm 0 data mismatch")
	}
	if decoded.Alarms[1].AlarmType != AlarmTypeTamper {
		t.Errorf("alarm 1 type: expected 0x06, got 0x%02x", decoded.Alarms[1].AlarmType)
	}
	t.Logf("Alarm info: %d bytes, %d alarms", len(data), len(decoded.Alarms))
}

func TestControlCode(t *testing.T) {
	code := &ControlCode{
		ControlType: ControlTypeSet,
		SerialNum:   100,
		ParamCount:  2,
		Params: []ControlParam{
			{ParamID: 0x1234, ParamLen: 2, ParamValue: []byte{0xAB, 0xCD}},
			{ParamID: 0x5678, ParamLen: 4, ParamValue: []byte{0x01, 0x02, 0x03, 0x04}},
		},
	}

	data, err := code.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded ControlCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.SerialNum != 100 {
		t.Errorf("SerialNum: expected 100, got %d", decoded.SerialNum)
	}
	if len(decoded.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(decoded.Params))
	}
	if decoded.Params[0].ParamID != 0x1234 {
		t.Errorf("param 0 ID mismatch")
	}
	if decoded.Params[1].ParamID != 0x5678 {
		t.Errorf("param 1 ID mismatch")
	}
	t.Logf("Control code: %d bytes, %d params", len(data), len(decoded.Params))
}

func TestControlResponseCode(t *testing.T) {
	resp := &ControlResponseCode{
		SerialNum:   100,
		ControlType: ControlTypeSet,
		Result:      0x01,
		DataLen:     4,
		Data:        []byte{0x0A, 0x0B, 0x0C, 0x0D},
	}

	data, err := resp.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded ControlResponseCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.SerialNum != 100 {
		t.Errorf("SerialNum mismatch")
	}
	if decoded.Result != 0x01 {
		t.Errorf("Result: expected 0x01, got 0x%02x", decoded.Result)
	}
	if !bytesEqual(decoded.Data, []byte{0x0A, 0x0B, 0x0C, 0x0D}) {
		t.Errorf("response data mismatch")
	}
	t.Logf("Control response: %d bytes", len(data))
}

func TestFileUploadNotification(t *testing.T) {
	notif := &FileUploadNotificationCode{
		FileNameLen: 10,
		FileName:    []byte("firmware.bin"),
		FileSize:    1048576,
		TotalBlocks: 256,
	}

	data, err := notif.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded FileUploadNotificationCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.FileSize != 1048576 {
		t.Errorf("FileSize: expected 1048576, got %d", decoded.FileSize)
	}
	if decoded.TotalBlocks != 256 {
		t.Errorf("TotalBlocks: expected 256, got %d", decoded.TotalBlocks)
	}
	if !bytesEqual(decoded.FileName, []byte("firmware.bin")) {
		t.Errorf("FileName mismatch: got %q", decoded.FileName)
	}
	t.Logf("File notification: %d bytes, file=%q size=%d blocks=%d",
		len(data), decoded.FileName, decoded.FileSize, decoded.TotalBlocks)
}

func TestFileDataBlock(t *testing.T) {
	block := &FileDataBlockCode{
		BlockIndex: 42,
		BlockLen:   1024,
		BlockData:  make([]byte, 1024),
	}
	for i := range block.BlockData {
		block.BlockData[i] = byte(i % 256)
	}

	data, err := block.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded FileDataBlockCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.BlockIndex != 42 {
		t.Errorf("BlockIndex mismatch")
	}
	if len(decoded.BlockData) != 1024 {
		t.Errorf("BlockData len: expected 1024, got %d", len(decoded.BlockData))
	}
	t.Logf("File data block: %d bytes, index=%d len=%d", len(data), decoded.BlockIndex, len(decoded.BlockData))
}

func TestFileUploadComplete(t *testing.T) {
	complete := &FileUploadCompleteCode{
		Result:    0x01,
		BlockMask: []byte{0x00, 0x00, 0x00, 0x00},
	}

	data, err := complete.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var decoded FileUploadCompleteCode
	if err := decoded.Decode(data); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.Result != 0x01 {
		t.Errorf("Result: expected 0x01, got 0x%02x", decoded.Result)
	}
	t.Logf("File upload complete: %d bytes, result=%d", len(data), decoded.Result)
}
