package model_test

import (
	"encoding/binary"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/im10furry/hj1239-go-sdk/model"
	"github.com/im10furry/hj1239-go-sdk/packet"
)

// ============================================================
// Simulation Test 1: Vehicle Terminal Full Cycle
// ============================================================

func TestSimulateVehicleTerminalFullCycle(t *testing.T) {
	vin := "LFV2A21K3G1234567"
	login := &model.VehicleLoginCode{
		Time:      model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0},
		SerialNum: 1,
		ICCID:     "89860000000000000001",
		AuthData:  []byte(vin),
		LoginTime: model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 1},
	}
	loginRaw, err := login.Encode()
	if err != nil {
		t.Fatalf("login encode: %v", err)
	}
	t.Logf("[SIM] Vehicle login encoded: %d bytes, VIN=%s", len(loginRaw), vin)

	var loginBack model.VehicleLoginCode
	if err := loginBack.Decode(loginRaw); err != nil {
		t.Fatalf("login decode: %v", err)
	}
	if loginBack.ICCID != login.ICCID {
		t.Errorf("ICCID mismatch: got %q, want %q", loginBack.ICCID, login.ICCID)
	}

	serial := uint16(1)
	for i := 0; i < 5; i++ {
		speed := 50.0 + float64(i)*10.0
		rpm := 1200.0 + float64(i)*200.0
		lon := 116.3912 + float64(i)*0.01
		lat := 39.9073 + float64(i)*0.005

		engine := model.NewEngineDataDPFSCR(
			speed, rpm, 101.3, 15.0+float64(i)*2.0,
			300.0+float64(i)*10, 150.0+float64(i)*8,
			280.0, 250.0, 3.0, 85.0, 75.0-float64(i)*2,
			40.0, 15.0, 120.0,
			lon, lat, 50000.0+float64(i)*10,
			true, true,
		)

		engineRaw, _ := engine.Encode()
		block := model.NewRealTimeInfoBlock(0x02, model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: uint8(i), Second: 30}, engineRaw)

		report := &model.RealTimeInfoReport{
			DataTime:   model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: uint8(i), Second: 31},
			SerialNum:  serial,
			InfoBlocks: []model.RealTimeInfoBlock{block},
		}
		serial++

		rptRaw, err := report.Encode()
		if err != nil {
			t.Fatalf("report %d encode: %v", i+1, err)
		}

		dataTime, _ := model.ParseGB1239Time(rptRaw[0:6])
		decSerial := binary.BigEndian.Uint16(rptRaw[6:8])
		t.Logf("[SIM] Report #%d: %d bytes, time=%02d:%02d:%02d, serial=%d, speed=%.1f km/h, rpm=%.0f",
			i+1, len(rptRaw),
			dataTime.Hour, dataTime.Minute, dataTime.Second,
			decSerial, speed, rpm)
	}

	logout := model.NewVehicleLogoutCode(serial)
	logout.Time = model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 5, Second: 0}
	logoutRaw, err := logout.Encode()
	if err != nil {
		t.Fatalf("logout encode: %v", err)
	}
	if len(logoutRaw) != 8 {
		t.Errorf("logout size: expected 8, got %d", len(logoutRaw))
	}
	t.Logf("[SIM] Vehicle logout: serial=%d, size=%d bytes", logout.SerialNum, len(logoutRaw))
}

// ============================================================
// Simulation Test 2: Enterprise Platform Full Cycle
// ============================================================

func TestSimulateEnterprisePlatformCycle(t *testing.T) {
	vin := "LFV2A21K3G1234567"
	username := "ent_plat_01"
	password := "secure_password_123"

	loginData := model.NewPlatformLoginCode(username, password, 0x01)
	loginRaw, _ := loginData.Encode()
	t.Logf("[SIM] Platform login: %d bytes, user=%s", len(loginRaw), username)

	pkt, _ := packet.BuildPlatformPacket(0x07, 0xFE, vin, 0x01, loginRaw)
	wire, _ := packet.EncodePlatformPacket(pkt)
	t.Logf("[SIM] Platform login packet: %d bytes on wire", len(wire))

	if !packet.VerifyChecksum(wire) {
		t.Error("platform login: checksum failed")
	}

	parsed, err := packet.ParsePlatformPacket(wire)
	if err != nil {
		t.Fatalf("parse platform packet: %v", err)
	}
	if parsed.CommandFlag != 0x07 {
		t.Errorf("command flag: expected 0x07, got 0x%02x", parsed.CommandFlag)
	}
	if parsed.VIN != vin {
		t.Errorf("VIN mismatch: %q != %q", parsed.VIN, vin)
	}

	decoded, _ := packet.DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
	decLogin := decoded.(*model.PlatformLoginCode)
	if decLogin.Username != username {
		t.Errorf("username: %q != %q", decLogin.Username, username)
	}

	for i := 0; i < 3; i++ {
		engine := model.NewEngineDataDPFSCR(
			60.0+float64(i)*5, 1500.0, 101.3, 18.0,
			350.0, 180.0, 260.0, 240.0, 4.5, 88.0, 72.0,
			35.0, 12.0, 120.0,
			116.3912, 39.9073, 51000.0+float64(i)*10,
			true, true,
		)
		engineRaw, _ := engine.Encode()

		fwdPkt, _ := packet.BuildPlatformPacket(0x02, 0xFE, vin, 0x01, engineRaw)
		fwdWire, _ := packet.EncodePlatformPacket(fwdPkt)

		if !packet.VerifyChecksum(fwdWire) {
			t.Errorf("forward #%d: checksum failed", i+1)
		}

		fwdParsed, _ := packet.ParsePlatformPacket(fwdWire)
		fwdDecoded, _ := packet.DecodeDataUnit(fwdParsed.CommandFlag, fwdParsed.DataUnit)
		_ = fwdDecoded

		t.Logf("[SIM] Platform forward #%d: %d bytes, cmd=0x%02x",
			i+1, len(fwdWire), fwdParsed.CommandFlag)
	}

	logoutData := model.NewPlatformLogoutCode(1)
	logoutRaw, _ := logoutData.Encode()
	logoutPkt, _ := packet.BuildPlatformPacket(0x08, 0xFE, vin, 0x01, logoutRaw)
	logoutWire, _ := packet.EncodePlatformPacket(logoutPkt)

	if !packet.VerifyChecksum(logoutWire) {
		t.Error("platform logout: checksum failed")
	}
	t.Logf("[SIM] Platform logout: %d bytes on wire", len(logoutWire))
}

// ============================================================
// Simulation Test 3: Packet Corruption & Recovery
// ============================================================

func TestSimulatePacketCorruptionRecovery(t *testing.T) {
	vin := "TESTVIN0000000001"
	collectTime := model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0}
	dataTime := model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 1}

	engine := model.NewEngineDataDPFSCR(
		60.0, 1500.0, 101.3, 18.0,
		350.0, 180.0, 260.0, 240.0, 4.5, 88.0, 75.0,
		35.0, 12.0, 120.0, 116.3912, 39.9073, 51234.5,
		true, true,
	)
	engineRaw, _ := engine.Encode()

	report := &model.RealTimeInfoReport{
		DataTime:  dataTime,
		SerialNum: 1,
		InfoBlocks: []model.RealTimeInfoBlock{
			model.NewRealTimeInfoBlock(0x02, collectTime, engineRaw),
		},
	}
	reportRaw, _ := report.Encode()

	origPkt, _ := packet.BuildPlatformPacket(0x02, 0xFE, vin, 0x01, reportRaw)
	origWire, _ := packet.EncodePlatformPacket(origPkt)

	if !packet.VerifyChecksum(origWire) {
		t.Fatal("original packet must pass checksum")
	}

	corrupted := make([]byte, len(origWire))
	copy(corrupted, origWire)
	flipPos := 25
	corrupted[flipPos] ^= 0x01

	if packet.VerifyChecksum(corrupted) {
		t.Error("corrupted packet should fail checksum")
	}

	err := packet.CheckChecksum(corrupted)
	if err == nil {
		t.Error("CheckChecksum should return error for corruption")
	} else {
		t.Logf("[SIM] Corruption detected: %v", err)
	}

	tampered := make([]byte, len(origWire))
	copy(tampered, origWire)
	tampered[4] = 'X'

	if packet.VerifyChecksum(tampered) {
		t.Error("tampered VIN should fail checksum")
	}

	retransmit := make([]byte, len(origWire))
	copy(retransmit, origWire)
	if !packet.VerifyChecksum(retransmit) {
		t.Error("retransmitted intact packet must pass checksum")
	}

	parsed, err := packet.ParsePlatformPacket(retransmit)
	if err != nil {
		t.Fatalf("parse retransmit: %v", err)
	}
	decoded, err := packet.DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
	if err != nil {
		t.Fatalf("decode retransmit: %v", err)
	}
	decReport := decoded.(*model.RealTimeInfoReport)
	eBody, err := model.ParseInfoBody(0x02, decReport.InfoBlocks[0].BodyRaw())
	if err != nil {
		t.Fatalf("parse info body: %v", err)
	}
	e := eBody.(*model.EngineDataDPFSCR)
	if !e.Speed.Valid {
		t.Error("retransmitted data: speed should be valid")
	}
	t.Logf("[SIM] Recovery successful: retransmitted packet parsed, checksum OK")
}

// ============================================================
// Simulation Test 4: Multi-Vehicle Concurrent Data Stream
// ============================================================

func TestSimulateMultiVehicleConcurrent(t *testing.T) {
	vins := []string{
		"VIN00000000000001",
		"VIN00000000000002",
		"VIN00000000000003",
		"VIN00000000000004",
		"VIN00000000000005",
	}

	const reportsPerVehicle = 20
	type vehicleReport struct {
		vin    string
		serial int
		raw    []byte
	}

	results := make(chan vehicleReport, len(vins)*reportsPerVehicle)
	var wg sync.WaitGroup

	for _, vin := range vins {
		wg.Add(1)
		go func(vehicleVIN string) {
			defer wg.Done()
			for i := 0; i < reportsPerVehicle; i++ {
				engine := model.NewEngineDataDPFSCR(
					float64(40+i), 1200.0+float64(i)*50, 101.3, 15.0+float64(i)*2,
					400.0-float64(i), 200.0-float64(i)*2,
					250.0, 220.0, 3.0, 90.0, 80.0-float64(i),
					30.0, 10.0, 120.0,
					116.0+float64(i)*0.1, 39.0+float64(i)*0.05,
					float64(50000+i*100),
					true, true,
				)
				engineRaw, _ := engine.Encode()
				pkt, _ := packet.BuildPlatformPacket(0x02, 0xFE, vehicleVIN, 0x01, engineRaw)
				wire, _ := packet.EncodePlatformPacket(pkt)

				results <- vehicleReport{
					vin:    vehicleVIN,
					serial: i,
					raw:    wire,
				}
			}
		}(vin)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	countByVIN := make(map[string]int)
	for r := range results {
		countByVIN[r.vin]++

		parsed, err := packet.ParsePlatformPacket(r.raw)
		if err != nil {
			t.Errorf("vehicle %s report #%d: parse failed: %v", r.vin, r.serial, err)
			continue
		}

		if !packet.VerifyChecksum(r.raw) {
			t.Errorf("vehicle %s report #%d: checksum failed", r.vin, r.serial)
			continue
		}

		if parsed.VIN != r.vin {
			t.Errorf("vehicle %s report #%d: VIN mismatch %q", r.vin, r.serial, parsed.VIN)
		}

		if _, err := packet.DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit); err != nil {
			t.Errorf("vehicle %s report #%d: decode failed: %v", r.vin, r.serial, err)
		}
	}

	for vin, count := range countByVIN {
		if count != reportsPerVehicle {
			t.Errorf("vehicle %s: expected %d reports, got %d", vin, reportsPerVehicle, count)
		}
	}
	t.Logf("[SIM] Multi-vehicle: %d vehicles x %d reports = %d total packets processed",
		len(vins), reportsPerVehicle, len(vins)*reportsPerVehicle)
}

// ============================================================
// Simulation Test 5: Edge Cases & Boundary Values
// ============================================================

func TestSimulateEdgeCases(t *testing.T) {
	t.Run("all_invalid_fields", func(t *testing.T) {
		invalid := make([]byte, model.EngineDataDPFSCRSize)
		for i := range invalid {
			invalid[i] = 0xFF
		}
		var e model.EngineDataDPFSCR
		if err := e.Decode(invalid); err != nil {
			t.Fatalf("decode all-invalid: %v", err)
		}
		if e.Speed.Valid {
			t.Error("speed should be invalid (0xFFFF)")
		}
		if e.Odometer.Valid {
			t.Error("odometer should be invalid (0xFFFFFFFF)")
		}
		t.Logf("[SIM] All invalid fields: speed.Valid=%v, odometer.Valid=%v",
			e.Speed.Valid, e.Odometer.Valid)
	})

	t.Run("boundary_min_max", func(t *testing.T) {
		eMin := model.NewEngineDataDPFSCR(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, true, true)
		rawMin, _ := eMin.Encode()
		var decMin model.EngineDataDPFSCR
		decMin.Decode(rawMin)
		if !decMin.Speed.Valid {
			t.Error("min speed should be valid")
		}
		if decMin.EngineSpeed.Raw != 0 {
			t.Errorf("min rpm raw: expected 0, got %d", decMin.EngineSpeed.Raw)
		}

		eMax := &model.EngineDataDPFSCR{
			Speed:             model.ScaledUint16{Raw: 0xFFFE, Valid: true},
			AtmPressure:       model.ScaledUint8{Raw: 0xFE, Valid: true},
			EngineTorque:      model.ScaledInt8{Raw: 0xFE, Valid: true},
			FrictionTorque:    model.ScaledInt8{Raw: 0xFE, Valid: true},
			EngineSpeed:       model.ScaledUint16{Raw: 0xFFFE, Valid: true},
			FuelFlow:          model.ScaledUint16{Raw: 0xFFFE, Valid: true},
			SCRUpstreamNOx:    model.ScaledInt16{Raw: 0xFFFE, Valid: true},
			SCRDownstreamNOx:  model.ScaledInt16{Raw: 0xFFFE, Valid: true},
			ReagentRate:       model.ScaledUint8{Raw: 0xFE, Valid: true},
			AirFlow:           model.ScaledUint16{Raw: 0xFFFE, Valid: true},
			SCRUpstreamTemp:   model.ScaledInt16{Raw: 0xFFFE, Valid: true},
			SCRDownstreamTemp: model.ScaledInt16{Raw: 0xFFFE, Valid: true},
			DPFDiffPressure:   model.ScaledUint16{Raw: 0xFFFE, Valid: true},
			CoolantTemp:       model.ScaledInt8{Raw: 0xFE, Valid: true},
			UreaLevel:         model.ScaledUint8{Raw: 0xFE, Valid: true},
			PositionStatus:    0x07,
			Longitude:         model.ScaledUint32{Raw: 0xFFFFFFFE, Valid: true},
			Latitude:          model.ScaledUint32{Raw: 0xFFFFFFFE, Valid: true},
			Odometer:          model.ScaledUint32{Raw: 0xFFFFFFFE, Valid: true},
		}
		rawMax, _ := eMax.Encode()
		var decMax model.EngineDataDPFSCR
		decMax.Decode(rawMax)
		t.Logf("[SIM] Max values: speed.raw=0x%04x, longitude.raw=0x%08x",
			decMax.Speed.Raw, decMax.Longitude.Raw)
	})

	t.Run("NaN_and_invalid_handling", func(t *testing.T) {
		nanVal := model.NewScaledUint16(math.NaN(), 0.05, 0)
		if nanVal.Valid {
			t.Error("NaN should produce invalid scaled value")
		}

		buf := make([]byte, 2)
		nanVal.PutUint16(buf, true)
		if buf[0] != 0xFF || buf[1] != 0xFF {
			t.Errorf("NaN encode: expected FFFF, got %02x%02x", buf[0], buf[1])
		}

		var dec model.ScaledUint16
		dec.DecodeUint16(buf, true)
		if dec.Valid {
			t.Error("0xFFFF decode should be invalid")
		}
	})

	t.Run("empty_and_minimal_packets", func(t *testing.T) {
		pkt, _ := packet.BuildPlatformPacket(0x04, 0xFE, "VIN00000000000001", 0x01, nil)
		wire, _ := packet.EncodePlatformPacket(pkt)
		if !packet.VerifyChecksum(wire) {
			t.Error("empty data packet should have valid checksum")
		}
		parsed, _ := packet.ParsePlatformPacket(wire)
		if parsed.DataLength != 0 {
			t.Errorf("data length: expected 0, got %d", parsed.DataLength)
		}

		pkt2, _ := packet.BuildPlatformPacket(0x01, 0xFE, "VIN00000000000001", 0x01, []byte{0xAA})
		wire2, _ := packet.EncodePlatformPacket(pkt2)
		if !packet.VerifyChecksum(wire2) {
			t.Error("1-byte data packet should have valid checksum")
		}
		t.Logf("[SIM] Empty packet: %d bytes, 1-byte packet: %d bytes", len(wire), len(wire2))
	})

	t.Run("mixed_info_types", func(t *testing.T) {
		obd := &model.OBDInfoBody{MilStatus: 0x0001, DTCCount: 0, VIN: "VIN00000000000001"}
		obdRaw, _ := obd.Encode()

		engine := model.NewEngineDataDPFSCR(60, 1500, 101.3, 18, 350, 180, 260, 240, 4.5, 88, 75, 35, 12, 120.0, 116.39, 39.91, 50000, true, true)
		engineRaw, _ := engine.Encode()

		block1 := model.NewRealTimeInfoBlock(0x01, model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0}, obdRaw)
		block2 := model.NewRealTimeInfoBlock(0x02, model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 1}, engineRaw)

		report := &model.RealTimeInfoReport{
			DataTime:   model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 2},
			SerialNum:  1,
			InfoBlocks: []model.RealTimeInfoBlock{block1, block2},
		}
		rptRaw, _ := report.Encode()
		t.Logf("[SIM] Mixed report: %d bytes (%d blocks)", len(rptRaw), len(report.InfoBlocks))

		var decoded model.RealTimeInfoReport
		if err := decoded.Decode(rptRaw); err != nil {
			t.Fatalf("decode mixed report: %v", err)
		}
		if n := len(decoded.InfoBlocks); n != 2 {
			t.Errorf("expected 2 blocks, got %d", n)
		}
		t.Logf("[SIM] Mixed report decoded: %d blocks", len(decoded.InfoBlocks))
	})
}

// ============================================================
// Simulation Test 6: Multi-Message Type Round-Trip
// ============================================================

func TestSimulateAllCommandTypes(t *testing.T) {
	vin := "VINTEST1234567890"

	type cmdCase struct {
		cmd  byte
		name string
		data []byte
	}

	cases := []cmdCase{
		{0x01, "VehicleLogin", func() []byte {
			v, _ := model.NewVehicleLoginCode("89860000000000000001", vin).Encode()
			return v
		}()},
		{0x04, "VehicleLogout", func() []byte {
			v, _ := model.NewVehicleLogoutCode(1).Encode()
			return v
		}()},
		{0x05, "TimeCalibration", func() []byte {
			return model.GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0}.Bytes()
		}()},
		{0x06, "SupplVehicleInfo", func() []byte {
			info := &model.SupplementaryVehicleInfo{
				TorqueMode:      model.ScaledUint8{Raw: 1, Valid: true},
				AccelPedal:      model.ScaledUint8{Raw: 50, Valid: true},
				CumulativeFuel:  model.ScaledUint32{Raw: 10000, Valid: true},
				DPFUpstreamTemp: model.ScaledInt16{Raw: 8000, Valid: true},
			}
			v, _ := info.Encode()
			return v
		}()},
		{0x07, "PlatformLogin", func() []byte {
			v, _ := model.NewPlatformLoginCode("admin", "pass123", 0x01).Encode()
			return v
		}()},
		{0x08, "PlatformLogout", func() []byte {
			v, _ := model.NewPlatformLogoutCode(1).Encode()
			return v
		}()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, _ := packet.BuildPlatformPacket(c.cmd, 0xFE, vin, 0x01, c.data)
			wire, _ := packet.EncodePlatformPacket(p)

			if !packet.VerifyChecksum(wire) {
				t.Fatalf("checksum failed")
			}

			parsed, err := packet.ParsePlatformPacket(wire)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			if parsed.CommandFlag != c.cmd {
				t.Errorf("cmd flag: 0x%02x != 0x%02x", parsed.CommandFlag, c.cmd)
			}

			decoded, err := packet.DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}

			t.Logf("[SIM] %s: %d bytes wire, type=%T", c.name, len(wire), decoded)
		})
	}
}

// ============================================================
// Simulation Test 7: Clock Override for Deterministic Testing
// ============================================================

func TestSimulateWithFixedClock(t *testing.T) {
	origClock := model.Clock
	model.Clock = func() time.Time {
		return time.Date(2024, 3, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600))
	}
	defer func() { model.Clock = origClock }()

	login := model.NewVehicleLoginCode("ICCID001", "VIN001")
	if login.Time.Year != 24 || login.Time.Hour != 8 {
		t.Errorf("fixed clock: expected year=24 hour=8, got year=%d hour=%d",
			login.Time.Year, login.Time.Hour)
	}

	platformLogin := model.NewPlatformLoginCode("admin", "pass", 0x01)
	if platformLogin.AccessTime.Year != 24 {
		t.Errorf("fixed clock platform: expected year=24, got year=%d",
			platformLogin.AccessTime.Year)
	}

	t.Logf("[SIM] Fixed clock: login time=%02d:%02d:%02d",
		login.Time.Hour, login.Time.Minute, login.Time.Second)
}

// ============================================================
// Benchmark: Full pipeline encode -> packet -> decode
// ============================================================

func BenchmarkFullPipeline(b *testing.B) {
	vin := "BENCHVIN000000001"
	e := model.NewEngineDataDPFSCR(60, 1500, 101.3, 18, 350, 180, 260, 240, 4.5, 88, 75, 35, 12, 120.0, 116.39, 39.91, 50000, true, true)
	engineRaw, _ := e.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := packet.BuildPlatformPacket(0x02, 0xFE, vin, 0x01, engineRaw)
		wire, _ := packet.EncodePlatformPacket(p)
		_ = packet.VerifyChecksum(wire)
		parsed, _ := packet.ParsePlatformPacket(wire)
		_, _ = packet.DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
	}
}
