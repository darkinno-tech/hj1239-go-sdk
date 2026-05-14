# hj1239-go-sdk

> *"我们是 DarkInno。如同浓烈世涛啤酒，我们最好的想法在黑暗中缓慢酝酿，远离喧嚣。"*

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](./LICENSE)
[![Standard](https://img.shields.io/badge/标准-HJ%201239.3--2021-blue)](https://www.mee.gov.cn/)

**HJ 1239.3-2021**《重型车排放远程监控技术规范 第3部分：通讯协议及数据格式》的 Go 语言 SDK。

## 概述

本 SDK 实现了 HJ 1239.3-2021 定义的二进制通讯协议，覆盖车载终端（第4章）和企业平台（第5章）的全部数据格式。

- **通讯协议**: TCP/IP，大端字节序，BCC 异或校验
- **标准依据**: HJ 1239.3-2021，引用 GB 17691-2018 附录 Q

## 安装

```bash
go get github.com/DarkInno/hj1239-go-sdk
```

要求 Go 1.21+。

## 快速开始

### 1. 构建并编码平台数据包

```go
import (
    "github.com/DarkInno/hj1239-go-sdk/model"
    "github.com/DarkInno/hj1239-go-sdk/packet"
)

// 构建企业平台登录包
p, _ := packet.BuildPlatformPacket(
    0x07,                   // 企业平台登录
    0xFE,                   // 请求包
    "VIN12345678901234",    // VIN
    0x01,                   // 不加密
    loginBytes,
)
data, _ := packet.EncodePlatformPacket(p)
// 通过 TCP 发送...
```

### 2. 解析平台数据包

```go
p, err := packet.ParsePlatformPacket(rawBytes)
if !packet.VerifyChecksum(rawBytes) {
    // 校验失败
}
decoded, _ := packet.DecodeDataUnit(p.CommandFlag, p.DataUnit)
switch v := decoded.(type) {
case *model.VehicleLoginCode:
    fmt.Printf("车辆登入: VIN=%s, ICCID=%s\n", p.VIN, v.ICCID)
}
```

### 3. 发动机数据 (DPF/SCR) 编解码

```go
engine := model.NewEngineDataDPFSCR(
    60.0, 1500.0, 101.3, 18.0,   // 车速, 转速, 大气压, 燃油流量
    350.0, 180.0, 260.0, 240.0,   // SCR NOx, 温度
    4.5, 88.0, 75.0,              // DPF压差, 冷却液温度, 尿素液位
    35.0, 12.0, 120.0,            // 扭矩百分比, 进气量
    116.3912, 39.9073, 51234.5,   // 经度, 纬度, 累计里程
    true, true,                   // 定位有效, 北纬
)

data, _ := engine.Encode()
// data 为 37 字节二进制数据 (Table 5)
```

## 包结构

| 包 | 说明 |
|----|------|
| `model` | 数据类型定义：车载终端/企业平台/附录A/B |
| `codec` | 二进制编解码工具：字节序、缩放、BCC校验 |
| `packet` | 数据包组帧/解帧 (5.6.3节)，命令分发 |
| `cmd` | 命令常量定义 |
| `gen` | 代码生成器，从 struct tag 生成 Marshal/Unmarshal |

## 已实现数据类型

### 第4章 — 车载终端

| 类型 | 对应表 | 大小 |
|------|--------|------|
| `VehicleLoginCode` | GB17691 Q.6.4.5.1 | 可变 |
| `VehicleLogoutCode` | GB17691 Q.6.4.5.3 | 8 字节 |
| `TerminalTimeCalibrationCode` | GB17691 Q.6.4.5.4 | 6 字节 |
| `EngineDataDPFSCR` | 表5 | 37 字节 |
| `EngineDataTWC` | 表6 | 30 字节 |
| `EngineDataTWCNOx` | 表7 | 32 字节 |
| `EngineDataHybrid` | 表8 | 8 字节 |
| `RealTimeInfoReport` | 表2 | 可变 |
| `VehicleInfoCode` | 表12 | 可变 |
| `VehicleInfoResponseCode` | 表13 | 2 字节 |
| `TerminalSupplementCode` | 附录A | 17 字节 |

### 第5章 — 企业平台

| 类型 | 对应表 | 大小 |
|------|--------|------|
| `PlatformPacket` | 表16 | 24 + data |
| `PlatformLoginCode` | 表19 | 41 字节 |
| `PlatformLogoutCode` | 表20 | 8 字节 |
| `KeyExchangeCode` | 表21 | 可变 |

## 缩放类型

协议使用缩放编码：`实际值 = 原始值 × 缩放因子 + 偏移量`

| 类型 | 原始类型 | 无效标记 |
|------|----------|----------|
| `ScaledUint8` | `uint8` | `0xFF` |
| `ScaledInt8` | `int8` | `0xFF` |
| `ScaledUint16` | `uint16` | `0xFFFF` |
| `ScaledInt16` | `int16` | `0xFFFF` |
| `ScaledUint32` | `uint32` | `0xFFFFFFFF` |

## 测试数据

### 单元测试 (38 项, 0 失败)

| 包 | 测试数 | 内容 |
|----|--------|------|
| `codec` | 7 | 字节读写、uint16/32、字符串、BCC、缩放、越界 |
| `model` | 23 | 时间、缩放、发动机数据、OBD、MIL、错误类型、7个模拟场景 |
| `packet` | 8 | 平台包编解码、校验、损坏检测 |

### 模拟场景

| 场景 | 数据包 | 结果 |
|------|--------|------|
| 车辆终端完整周期：登入→5次上报→登出 | 7 | ✅ |
| 企业平台周期：登录→3次转发→登出 | 5 | ✅ |
| 数据包损坏→校验失败→重传恢复 | 3 | ✅ |
| 多车并发：5车×20次上报 | 100 | ✅ |
| 边界值：全无效/最大最小值/NaN/空包/混合类型 | 6 | ✅ |
| 全部6种命令类型往返 | 6 | ✅ |
| 固定时钟确定性测试 | 2 | ✅ |

### 压力测试

| 测试 | 规模 | 吞吐量 | 错误 |
|------|------|--------|------|
| 顺序 100K | 100,000 包 | 294 万 pkt/s (174 MiB/s) | 0 |
| 并发 16×3125 | 50,000 包 | 386 万 pkt/s (228 MiB/s) | 0 |
| 全部命令 ×7 | 20,000 包 | 657 万 pkt/s | 0 |
| 损坏检测 | 10,000 包 | 100.0% 检出 | 0 |
| 延迟分布 | 5,000 次 | P50=0ns P99=0ns Max=879µs | 0 |

### 基准性能

```
BenchmarkEngineDataDPFSCREncode    47M ops/s    24.2 ns/op
BenchmarkEngineDataDPFSCRDecode   186M ops/s     6.0 ns/op
BenchmarkFullPipeline              4.6M ops/s   238.6 ns/op
```

### 协议合规

| 项目 | 状态 |
|------|------|
| 字节序：大端 (5.6.2节) | ✅ |
| BCC 异或校验 (5.6.3节) | ✅ |
| 表5: DPF/SCR 发动机数据 | ✅ |
| 表6-8: TWC/TWC+NOx/混合动力 | ✅ |
| 表10: 定位状态位 | ✅ |
| 表12/13: 车辆信息/应答 | ✅ |
| 表16: 平台数据包结构 | ✅ |
| 表19-21: 平台登录/登出/密钥交换 | ✅ |
| 附录A+B: 补充数据 | ✅ |
| 全部缩放因子和偏移量 | ✅ |

### 已知限制

- 多信息块实时上报解析需明确每种类型大小
- 加密模式 (SM2/SM4/RSA/AES128) 已声明但未实现
- TCP 帧转义 (0x7E 字节填充) 留给传输层处理

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
