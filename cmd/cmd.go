package cmd

// 车辆终端命令单元 (上行)
const (
	CmdVehicleLogin           byte = 0x01 // 车辆登入
	CmdRealTimeInfo           byte = 0x02 // 实时信息上报
	CmdHistoricalInfo         byte = 0x03 // 历史信息上报
	CmdVehicleLogout          byte = 0x04 // 车辆登出
	CmdTerminalTimeCalib      byte = 0x05 // 终端校时 (下行)
	CmdSupplementaryInfo      byte = 0x06 // 补充车辆信息
	CmdVehicleInfo            byte = 0x07 // 车辆信息
	CmdVehicleInfoResponse    byte = 0x08 // 车辆信息应答
	CmdAlarmInfo              byte = 0x0A // 告警信息上报
	CmdFileUploadNotification byte = 0x0B // 文件上传通知
	CmdFileDataBlock          byte = 0x0C // 文件数据块
	CmdFileUploadComplete     byte = 0x0D // 文件上传完成
)

// 企业平台命令单元 (上行/下行)
const (
	CmdPlatformLogin  byte = 0x07 // 企业平台登录
	CmdPlatformLogout byte = 0x08 // 企业平台登出
	CmdKeyExchange    byte = 0x09 // 密钥交换
)

// 远程控制命令 (平台 → 终端下行)
const (
	CmdControlRequest  byte = 0x81 // 远程控制请求
	CmdControlResponse byte = 0x82 // 远程控制应答
)

// 响应标志
const (
	RespSuccess     byte = 0x01 // 成功
	RespFailure     byte = 0x02 // 失败
	RespVINDuplicate byte = 0x03 // VIN重复
	RespRequest     byte = 0xFE // 此包为请求包
)

// 加密方式
const (
	EncryptNone    byte = 0x01 // 不加密
	EncryptSM2     byte = 0x02 // SM2 算法
	EncryptSM4     byte = 0x03 // SM4 算法
	EncryptRSA     byte = 0x04 // RSA 算法
	EncryptAES128  byte = 0x05 // AES128 算法
	EncryptInvalid byte = 0xFF
)

// 信息类型标志 (Table 4)
const (
	InfoTypeOBD         byte = 0x01 // OBD信息
	InfoTypeEngineDPFSCR    byte = 0x02 // 发动机数据 (DPF/SCR)
	InfoTypeEngineTWC       byte = 0x03 // 发动机数据 (TWC)
	InfoTypeEngineHybrid    byte = 0x04 // 发动机数据 (混合动力)
	InfoTypeEngineTWCNOx    byte = 0x05 // 发动机数据 (TWC+NOx)
	InfoTypeSupplement      byte = 0x80 // 补充车辆信息
)

const (
	// PacketStartMarker 数据包起始标识
	PacketStartMarker = "\x7e\x7e"
)
