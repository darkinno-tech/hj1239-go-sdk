package model

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// Stress Test 1: 100K Sequential Packets
// ============================================================

func TestStressSequential100K(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const N = 100_000
	vin := "STRESSVIN00000001"
	var totalBytes int64
	var checksumErrors int64
	start := time.Now()

	for i := 0; i < N; i++ {
		engine := NewEngineDataDPFSCR(
			float64(i%120), float64(1000+i%3000), 101.3, 15.0+float64(i%30),
			400.0-float64(i%100), 200.0-float64(i%50),
			250.0, 220.0, 3.0+float64(i%5), 85.0+float64(i%20), 80.0,
			30.0, 10.0, float64(100+i%200),
			116.0+float64(i%180)*0.0001, 39.0+float64(i%90)*0.0001,
			float64(50000+i),
			i%2 == 0, true,
		)
		engineRaw, _ := engine.Encode()

		pkt, _ := BuildPlatformPacket(0x02, 0xFE, vin, 0x01, engineRaw)
		wire, _ := EncodePlatformPacket(pkt)

		if !VerifyChecksum(wire) {
			checksumErrors++
		}

		parsed, _ := ParsePlatformPacket(wire)
		decoded, _ := DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
		_ = decoded

		atomic.AddInt64(&totalBytes, int64(len(wire)))
	}

	elapsed := time.Since(start)
	throughput := float64(N) / elapsed.Seconds()
	mibps := float64(atomic.LoadInt64(&totalBytes)) / elapsed.Seconds() / 1024 / 1024

	t.Logf("[STRESS] Sequential 100K: %d packets in %v (%.0f pkt/s, %.2f MiB/s) errors=%d",
		N, elapsed.Round(time.Millisecond), throughput, mibps, checksumErrors)

	if checksumErrors > 0 {
		t.Errorf("checksum errors: %d", checksumErrors)
	}
}

// ============================================================
// Stress Test 2: Concurrent 50K Packets (16 goroutines)
// ============================================================

func TestStressConcurrent50K(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const N = 50_000
	const workers = 16
	perWorker := N / workers

	var checksumErrors int64
	var totalBytes int64
	var wg sync.WaitGroup
	start := time.Now()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(workerID * 1234567)))
			vin := fmt.Sprintf("C%02dVIN%010d", workerID, workerID)

			for i := 0; i < perWorker; i++ {
				engine := NewEngineDataDPFSCR(
					rng.Float64()*120, rng.Float64()*4000+800, 101.3, rng.Float64()*40,
					rng.Float64()*500, rng.Float64()*300,
					rng.Float64()*400, rng.Float64()*400, rng.Float64()*10, rng.Float64()*30+70, rng.Float64()*100,
					rng.Float64()*80, rng.Float64()*30, rng.Float64()*300,
					rng.Float64()*180, rng.Float64()*90,
					rng.Float64()*200000+50000,
					i%2 == 0, true,
				)
				engineRaw, _ := engine.Encode()

				pkt, _ := BuildPlatformPacket(0x02, 0xFE, vin, 0x01, engineRaw)
				wire, _ := EncodePlatformPacket(pkt)

				if !VerifyChecksum(wire) {
					atomic.AddInt64(&checksumErrors, 1)
				}

				parsed, err := ParsePlatformPacket(wire)
				if err != nil {
					atomic.AddInt64(&checksumErrors, 1)
					continue
				}

				if parsed.VIN != vin {
					atomic.AddInt64(&checksumErrors, 1)
					continue
				}

				_, err = DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
				if err != nil {
					atomic.AddInt64(&checksumErrors, 1)
				}

				atomic.AddInt64(&totalBytes, int64(len(wire)))
			}
		}(w)
	}

	wg.Wait()
	elapsed := time.Since(start)
	throughput := float64(N) / elapsed.Seconds()
	mibps := float64(atomic.LoadInt64(&totalBytes)) / elapsed.Seconds() / 1024 / 1024

	t.Logf("[STRESS] Concurrent %dx%d: %d packets in %v (%.0f pkt/s, %.2f MiB/s) errors=%d",
		workers, perWorker, N, elapsed.Round(time.Millisecond), throughput, mibps, checksumErrors)

	if checksumErrors > 0 {
		t.Errorf("checksum errors: %d", checksumErrors)
	}
}

// ============================================================
// Stress Test 3: All Command Types Under Load
// ============================================================

func TestStressAllCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const N = 20000
	var wg sync.WaitGroup
	var errors int64
	start := time.Now()

	commands := []struct {
		cmd  byte
		name string
	}{
		{0x01, "Login"}, {0x02, "Realtime"}, {0x04, "Logout"},
		{0x05, "TimeCal"}, {0x06, "Suppl"}, {0x07, "PlatLogin"},
		{0x08, "PlatLogout"},
	}

	for _, c := range commands {
		wg.Add(1)
		go func(cmd byte, name string) {
			defer wg.Done()
			vin := fmt.Sprintf("CMDVIN%08d%s", cmd, "0000")
			perCmd := N / len(commands)

			for i := 0; i < perCmd; i++ {
				var data []byte
				switch cmd {
				case 0x01:
					v := &VehicleLoginCode{
						Time:      GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0},
						SerialNum: uint16(i),
						ICCID:     fmt.Sprintf("ICCID%015d", i),
						AuthData:  []byte(vin),
						LoginTime: GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 1},
					}
					data, _ = v.Encode()
				case 0x02:
					e := NewEngineDataDPFSCR(60, 1500, 101.3, 18, 350, 180, 260, 240, 4.5, 88, 75, 35, 12, 120, 116.39, 39.91, 50000, true, true)
					data, _ = e.Encode()
				case 0x04:
					v := NewVehicleLogoutCode(uint16(i))
					data, _ = v.Encode()
				case 0x05:
					data = GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0}.Bytes()
				case 0x06:
					v := &SupplementaryVehicleInfo{
						TorqueMode:  ScaledUint8{Raw: 1, Valid: true},
						AccelPedal:  ScaledUint8{Raw: 50, Valid: true},
						DPFUpstreamTemp: ScaledInt16{Raw: 8000, Valid: true},
					}
					data, _ = v.Encode()
				case 0x07:
					v := NewPlatformLoginCode(fmt.Sprintf("user%04d", i%9999), "pass", 0x01)
					data, _ = v.Encode()
				case 0x08:
					v := NewPlatformLogoutCode(uint16(i))
					data, _ = v.Encode()
				}

				pkt, _ := BuildPlatformPacket(cmd, 0xFE, vin, 0x01, data)
				wire, _ := EncodePlatformPacket(pkt)

				if !VerifyChecksum(wire) {
					atomic.AddInt64(&errors, 1)
					continue
				}

				parsed, _ := ParsePlatformPacket(wire)
				_, err := DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
				if err != nil {
					atomic.AddInt64(&errors, 1)
				}
			}
		}(c.cmd, c.name)
	}

	wg.Wait()
	elapsed := time.Since(start)
	t.Logf("[STRESS] All commands: %d packets in %v (%.0f pkt/s) errors=%d",
		N, elapsed.Round(time.Millisecond), float64(N)/elapsed.Seconds(), errors)

	if errors > 0 {
		t.Errorf("total errors: %d", errors)
	}
}

// ============================================================
// Stress Test 4: Memory Allocation Pressure
// ============================================================

func TestStressMemoryAlloc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const N = 10000
	var allocBefore, allocAfter float64

	// Measure baseline
	testing.AllocsPerRun(1, func() {})

	avg := testing.AllocsPerRun(N, func() {
		e := NewEngineDataDPFSCR(60, 1500, 101.3, 18, 350, 180, 260, 240, 4.5, 88, 75, 35, 12, 120, 116.39, 39.91, 50000, true, true)
		data, _ := e.Encode()
		pkt, _ := BuildPlatformPacket(0x02, 0xFE, "VINALLOC0000001", 0x01, data)
		wire, _ := EncodePlatformPacket(pkt)
		parsed, _ := ParsePlatformPacket(wire)
		_, _ = DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
	})

	_ = allocBefore
	_ = allocAfter
	t.Logf("[STRESS] Memory: %.1f allocs/op for full pipeline", avg)
}

// ============================================================
// Stress Test 5: Small Packet Flood (1-byte data units)
// ============================================================

func TestStressSmallPackets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const N = 30000
	const workers = 8
	var errors int64
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(wid int) {
			defer wg.Done()
			vin := fmt.Sprintf("SM%dVIN%08d", wid, wid)
			for i := 0; i < N/workers; i++ {
				pkt, _ := BuildPlatformPacket(0x05, 0xFE, vin, 0x01,
					GB1239Time{Year: 24, Month: 3, Day: 15, Hour: 8, Minute: 0, Second: 0}.Bytes())
				wire, _ := EncodePlatformPacket(pkt)

				if !VerifyChecksum(wire) {
					atomic.AddInt64(&errors, 1)
					continue
				}

				parsed, err := ParsePlatformPacket(wire)
				if err != nil || parsed.DataLength != 6 {
					atomic.AddInt64(&errors, 1)
				}
			}
		}(w)
	}
	wg.Wait()
	t.Logf("[STRESS] Small packets: %d processed, errors=%d", N, errors)
	if errors > 0 {
		t.Errorf("small packet errors: %d", errors)
	}
}

// ============================================================
// Stress Test 6: Corruption Detection Rate
// ============================================================

func TestStressCorruptionDetection(t *testing.T) {
	const N = 10000
	var detected int64

	for i := 0; i < N; i++ {
		e := NewEngineDataDPFSCR(
			float64(i%120), float64(1000+i%3000), 101.3, 15,
			400, 200, 250, 220, 3, 85, 80, 30, 10, 120,
			116.39, 39.91, 50000+float64(i),
			true, true,
		)
		engineRaw, _ := e.Encode()
		pkt, _ := BuildPlatformPacket(0x02, 0xFE, "CORRVIN00000001", 0x01, engineRaw)
		wire, _ := EncodePlatformPacket(pkt)

		// Corrupt 1 random byte in payload
		corruptPos := 24 + (i % len(engineRaw))
		if corruptPos < len(wire)-1 {
			wire[corruptPos] ^= 0xFF
			if !VerifyChecksum(wire) {
				detected++
			}
		}
	}

	rate := float64(detected) / float64(N) * 100
	t.Logf("[STRESS] Corruption detection: %d/%d (%.1f%%)", detected, N, rate)
	if rate < 99.0 {
		t.Errorf("corruption detection rate too low: %.1f%%", rate)
	}
}

// ============================================================
// Stress Test 7: Latency Distribution
// ============================================================

func TestStressLatencyDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const N = 5000
	latencies := make([]time.Duration, N)
	e := NewEngineDataDPFSCR(60, 1500, 101.3, 18, 350, 180, 260, 240, 4.5, 88, 75, 35, 12, 120, 116.39, 39.91, 50000, true, true)
	engineRaw, _ := e.Encode()

	var maxLatency, minLatency time.Duration
	minLatency = time.Hour

	for i := 0; i < N; i++ {
		start := time.Now()
		pkt, _ := BuildPlatformPacket(0x02, 0xFE, "LATVIN000000001", 0x01, engineRaw)
		wire, _ := EncodePlatformPacket(pkt)
		_ = VerifyChecksum(wire)
		parsed, _ := ParsePlatformPacket(wire)
		_, _ = DecodeDataUnit(parsed.CommandFlag, parsed.DataUnit)
		lat := time.Since(start)
		latencies[i] = lat

		if lat > maxLatency {
			maxLatency = lat
		}
		if lat < minLatency {
			minLatency = lat
		}
	}

	// Calculate P50, P99
	// Simple: sort a copy (accept N=5000 is fine)
	sorted := make([]time.Duration, N)
	copy(sorted, latencies)

	// Insertion sort for small N
	for i := 1; i < N; i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	p50 := sorted[N/2]
	p99 := sorted[N*99/100]

	t.Logf("[STRESS] Latency (N=%d): min=%v p50=%v p99=%v max=%v",
		N, minLatency.Round(time.Nanosecond), p50.Round(time.Nanosecond),
		p99.Round(time.Nanosecond), maxLatency.Round(time.Nanosecond))
}

// Helper: non-cryptographic random source for stress tests
var stressRand = rand.New(rand.NewSource(time.Now().UnixNano()))
