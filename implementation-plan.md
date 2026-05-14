# GB1239 Go SDK — Implementation Plan

## 项目概览

- **名称**: hj1239-sdk-go
- **语言**: Go 1.21+
- **模块路径**: github.com/DarkInno/hj1239-go-sdk
- **核心业务**: HJ 1239.3-2021 重型车排放远程监控通讯协议及数据格式 SDK
- **模式**: 0→1 新建

## 项目背景

GB1239 是中国生态环境部发布的《重型车排放远程监控技术规范》第3部分（HJ 1239.3-2021），
定义了重型车排放远程监控的数据传输通讯协议及数据格式。

本 SDK 实现：
- 车辆终端通讯协议及数据格式（Section 4）
- 企业平台通讯协议及数据结构（Section 5）
- 附录 A：车辆终端补充数据
- 附录 B：补充车辆信息数据

## 架构设计

### 应用架构

- **模式**: Go 标准库（无框架依赖）
- **包结构**:
  - `model/` — 所有数据类型的 Go struct 定义
  - `codec/` — 二进制编解码框架（字节序、缩放、偏移量）
  - `packet/` — 数据包组帧/解帧（起始标识、校验、加密标识）
  - `gen/` — 代码生成器（基于 struct tag 生成 Marshal/Unmarshal）

### 分层

```
┌────────────────────────────────────────┐
│  public API (Marshal/Unmarshal)        │
├────────────────────────────────────────┤
│  packet/ — 包组帧/解帧                  │
├────────────────────────────────────────┤
│  codec/ — 二进制编解码引擎              │
├────────────────────────────────────────┤
│  model/ — 数据类型定义（含 tag 标记）     │
└────────────────────────────────────────┘
```

### 目录结构

```
hj1239-sdk-go/
├── model/
│   ├── types.go          # 基础类型和标签定义
│   ├── terminal.go       # 车载终端数据结构 (Section 4)
│   ├── platform.go       # 企业平台数据结构 (Section 5)
│   └── annex.go          # 附录 A/B 数据结构
├── codec/
│   ├── codec.go          # 编解码器接口和核心逻辑
│   ├── marshal.go        # MarshalBinary 实现
│   └── unmarshal.go      # UnmarshalBinary 实现
├── packet/
│   ├── packet.go         # 数据包结构 (Section 5.6.3)
│   └── checksum.go       # BCC 异或校验
├── cmd/
│   └── cmd.go            # 命令单元常量
├── gen/
│   ├── main.go           # go generate 入口
│   └── generator.go      # 代码生成逻辑
├── go.mod
├── go.sum
└── implementation-plan.md
```

### 技术选型

| 类型 | 选型 | 说明 |
|------|------|------|
| 语言 | Go 1.21+ | 泛型支持 |
| 编码 | encoding/binary | 大端字节序 |
| 测试 | testing + testify | 标准库测试 |
| 代码生成 | go generate + text/template | struct tag → codec |

### 协议要点

- **字节序**: 大端（Big-Endian）
- **校验**: BCC 异或校验
- **加密**: 支持 SM2/SM4/RSA/AES128
- **时间**: GMT+8，6 字节 (年/月/日/时/分/秒)

### 命令单元

| 命令号 | 名称 | 方向 |
|--------|------|------|
| 0x01 | 车辆登入 | 上行 |
| 0x02 | 实时信息上报 | 上行 |
| 0x03 | 历史信息上报 | 上行 |
| 0x04 | 车辆登出 | 上行 |
| 0x05 | 终端校时 | 下行 |
| 0x06 | 补充车辆信息 | 上行 |
| 0x07 | 车辆信息/企业平台登录 | 上行 |
| 0x08 | 车辆信息应答/企业平台登出 | 下行/上行 |
| 0x09 | 密钥交换 | 上行/下行 |

## 实施完成状态

### 已完成

| Phase | 内容 | 状态 |
|-------|------|------|
| Phase 1-2 | 项目分析与架构设计 | ✅ |
| Phase 3 | 数据模型定义 (model/) | ✅ |
| Phase 4 | 二进制编解码框架 (codec/) | ✅ |
| Phase 5 | 数据包组帧/解帧 (packet/) | ✅ |
| Phase 6 | 代码生成器 (gen/) | ✅ |
| Phase 7 | 测试与验证 (19 tests pass) | ✅ |
| Phase 8 | 文档输出 (README + examples) | ✅ |

### 文件清单

```
hj1239-sdk-go/
├── cmd/
│   ├── cmd.go              # 命令常量
│   └── gen/main.go         # 代码生成器 CLI
├── codec/
│   ├── codec.go            # 编解码核心
│   └── codec_test.go       # 编解码测试
├── gen/
│   └── generator.go        # 代码生成逻辑
├── model/
│   ├── types.go            # 基础类型 (GB1239Time, tag 定义)
│   ├── scaled.go           # 缩放数据类型 + 编解码辅助函数
│   ├── terminal.go         # 车载终端数据 (Section 4)
│   ├── platform.go         # 企业平台数据 (Section 5)
│   ├── annex.go            # 附录 A/B 数据
│   └── model_test.go       # 模型测试
├── packet/
│   ├── packet.go           # 数据包组帧/解帧 + 命令分发
│   └── packet_test.go      # 包处理测试
├── example/
│   └── example.go          # 使用示例
├── go.mod
├── implementation-plan.md
└── README.md
```

### 测试结果

```sh
$ go test ./... -count=1
ok      github.com/DarkInno/hj1239-go-sdk/codec    0.028s
ok      github.com/DarkInno/hj1239-go-sdk/model     0.030s
ok      github.com/DarkInno/hj1239-go-sdk/packet    0.024s
```
