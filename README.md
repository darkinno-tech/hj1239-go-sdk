# hj1239-go-sdk

> *"We are DarkInno. Like a stout beer, our best ideas are brewed slowly in the dark, away from the hype."*

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](./LICENSE)
[![Standard](https://img.shields.io/badge/Standard-HJ%201239.3--2021-blue)](https://www.mee.gov.cn/)

Go SDK for **HJ 1239.3-2021** (GB1239) — the Chinese national standard for
*Heavy-duty Vehicle Emission Remote Supervision System — Communication Protocol and Data Format*.

## Overview

This SDK implements the binary communication protocol defined in HJ 1239.3-2021,
covering both vehicle terminal (Section 4) and enterprise platform (Section 5) data formats.

**Protocol**: TCP/IP, Big-Endian byte order, BCC (XOR) checksum
**Standard**: HJ 1239.3-2021, referencing GB 17691-2018 Annex Q

## Installation

```bash
go get github.com/DarkInno/hj1239-go-sdk
```

Requires Go 1.21+.

## Quick Start

### 1. Build and Encode a Platform Packet

```go
package main

import (
    "fmt"
    "github.com/DarkInno/hj1239-go-sdk/model"
    "github.com/DarkInno/hj1239-go-sdk/packet"
)

func main() {
    p, _ := packet.BuildPlatformPacket(
        0x07,                   // 企业平台登录
        0xFE,                   // 请求包
        "VIN12345678901234",    // 车辆VIN
        0x01,                   // 不加密
        loginBytes,             // PlatformLoginCode encoded
    )
    data, _ := packet.EncodePlatformPacket(p)
    // Send data over TCP...
}
```

### 2. Decode a Platform Packet

```go
p, err := packet.ParsePlatformPacket(rawBytes)
if err != nil {
    // handle error
}
if !packet.VerifyChecksum(rawBytes) {
    // checksum mismatch
}
decoded, err := packet.DecodeDataUnit(p.CommandFlag, p.DataUnit)
switch v := decoded.(type) {
case *model.VehicleLoginCode:
    fmt.Printf("Vehicle login: VIN=%s, ICCID=%s\n", p.VIN, v.ICCID)
case *model.RealTimeInfoReport:
    fmt.Printf("Real-time info: %d blocks\n", len(v.InfoBlocks))
}
```

### 3. Vehicle Terminal — Engine Data (DPF/SCR)

```go
engine := &model.EngineDataDPFSCR{
    Speed:             model.ScaledUint16{Raw: 0x2000, Valid: true},  // 32 km/h
    EngineSpeed:       model.ScaledUint16{Raw: 0x1000, Valid: true},  // 512 rpm
    SCRUpstreamNOx:    model.ScaledInt16{Raw: 0x0500, Valid: true},
    SCRDownstreamNOx:  model.ScaledInt16{Raw: 0x0200, Valid: true},
    Longitude:         model.ScaledUint32{Raw: 116000000, Valid: true}, // 116.0°
    Latitude:          model.ScaledUint32{Raw: 40000000, Valid: true},  // 40.0°
}

data, err := engine.Encode()
// data is 37 bytes of binary data per Table 5

var decoded model.EngineDataDPFSCR
decoded.Decode(data)
```

## Package Structure

| Package | Description |
|---------|-------------|
| `model` | All data type definitions: vehicle terminal data, platform data, Annex A/B |
| `codec` | Binary encoding/decoding utilities: byte order, scaling, BCC |
| `packet` | Packet framing/deframing (Section 5.6.3), command dispatch |
| `cmd` | Command constant definitions |
| `gen` | Code generator for MarshalBinary/UnmarshalBinary from struct tags |

## Implemented Data Types

### Section 4 — Vehicle Terminal

| Type | Table | Size |
|------|-------|------|
| `VehicleLoginCode` | GB17691 Q.6.4.5.1 | Variable |
| `VehicleLogoutCode` | GB17691 Q.6.4.5.3 | 8 bytes |
| `TerminalTimeCalibrationCode` | GB17691 Q.6.4.5.4 | 6 bytes |
| `EngineDataDPFSCR` | Table 5 | 37 bytes |
| `EngineDataTWC` | Table 6 | 30 bytes |
| `EngineDataTWCNOx` | Table 7 | 32 bytes |
| `EngineDataHybrid` | Table 8 | 8 bytes |
| `RealTimeInfoReport` | Table 2 | Variable |
| `VehicleInfoCode` | Table 12 | Variable |
| `VehicleInfoResponseCode` | Table 13 | 2 bytes |
| `TerminalSupplementCode` | Annex A | 17 bytes |

### Section 5 — Enterprise Platform

| Type | Table | Size |
|------|-------|------|
| `PlatformPacket` | Table 16 | 24 + data |
| `PlatformLoginCode` | Table 19 | 41 bytes |
| `PlatformLogoutCode` | Table 20 | 8 bytes |
| `KeyExchangeCode` | Table 21 | Variable |

### Annex B — Supplementary Vehicle Info

| Type | Table | Size |
|------|-------|------|
| `SupplementaryVehicleInfo` | Table B.1 | Variable |

## Scaled Types

The SDK uses scaled types to handle the protocol's scale/offset encoding:

```
actual_value = raw_value * scale + offset_val
```

| Type | Go Type | Invalid Marker |
|------|---------|----------------|
| `ScaledUint8` | `uint8` | `0xFF` |
| `ScaledInt8` | `int8` (via `uint8`) | `0xFF` |
| `ScaledUint16` | `uint16` | `0xFFFF` |
| `ScaledInt16` | `int16` (via `uint16`) | `0xFFFF` |
| `ScaledUint32` | `uint32` | `0xFFFFFFFF` |

## Code Generation

Generate MarshalBinary/UnmarshalBinary methods from struct tags:

```bash
go run cmd/gen/main.go ./model

# Or via go:generate
//go:generate go run github.com/DarkInno/hj1239-go-sdk/cmd/gen ./model
```

## Test Data & Performance

### Unit Tests (38 tests, 0 failures)

```
ok  github.com/DarkInno/hj1239-go-sdk/codec    7 tests
ok  github.com/DarkInno/hj1239-go-sdk/model    23 tests
ok  github.com/DarkInno/hj1239-go-sdk/packet   8 tests
```

### Simulation Tests

| # | Scenario | Packets | Result |
|---|----------|---------|--------|
| 1 | Vehicle terminal login → 5x real-time → logout | 7 | ✅ |
| 2 | Platform login → 3x data forward → logout | 5 | ✅ |
| 3 | Packet corruption → detection → retransmit | 3 | ✅ |
| 4 | Multi-vehicle concurrent (5×20) | 100 | ✅ |
| 5 | Edge cases (invalid, boundary, NaN, empty, mixed) | 6 | ✅ |
| 6 | All 6 command types round-trip | 6 | ✅ |
| 7 | Fixed clock deterministic | 2 | ✅ |

### Stress Tests

| Test | Volume | Throughput | Errors |
|------|--------|------------|--------|
| Sequential 100K | 100,000 pkts | 2,943,835 pkt/s (174 MiB/s) | 0 |
| Concurrent 16×3125 | 50,000 pkts | 3,856,329 pkt/s (228 MiB/s) | 0 |
| All commands ×7 | 20,000 pkts | 6,566,635 pkt/s | 0 |
| Small packets (8 goroutines) | 30,000 pkts | — | 0 |
| Corruption detection | 10,000 corrupted | 100.0% detected | 0 |
| Latency distribution | 5,000 ops | P50=0ns P99=0ns Max=879µs | 0 |
| Memory allocation | 10,000 ops | 7.0 allocs/op | — |

### Micro-benchmarks

```
BenchmarkEngineDataDPFSCREncode    47M ops/s    24.2 ns/op
BenchmarkEngineDataDPFSCRDecode   186M ops/s     6.0 ns/op
BenchmarkFullPipeline (E→P→C→D)   4.6M ops/s   238.6 ns/op
```

### Protocol Compliance

| Item | Status |
|------|--------|
| Byte order: Big-Endian (Section 5.6.2) | ✅ |
| BCC XOR checksum (Section 5.6.3) | ✅ |
| Table 5: DPF/SCR engine data (37 bytes) | ✅ |
| Table 6: TWC engine data (30 bytes) | ✅ |
| Table 7: TWC+NOx engine data (32 bytes) | ✅ |
| Table 8: Hybrid data (8 bytes) | ✅ |
| Table 10: Positioning status bits | ✅ |
| Table 12/13: Vehicle info / response | ✅ |
| Table 16: Platform packet structure | ✅ |
| Table 19/20/21: Platform login/logout/key | ✅ |
| Annex A+B: Supplemental data | ✅ |
| All scale factors & offsets verified | ✅ |
| Invalid marker handling | ✅ |

### Known Limitations

- Multi-block `RealTimeInfoReport` body parsing requires explicit per-type body size
- Encryption (SM2/SM4/RSA/AES128) declared, not implemented — plaintext only
- TCP frame escaping (0x7E byte stuffing) left to transport layer

---

<p align="center">
  <a href="https://github.com/DarkInno/hj1239-go-sdk">
    <img src="https://img.shields.io/github/stars/DarkInno/hj1239-go-sdk?style=social" alt="GitHub stars">
  </a>
  &nbsp;
  <a href="https://github.com/DarkInno/hj1239-go-sdk/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-green" alt="MIT License">
  </a>
</p>
