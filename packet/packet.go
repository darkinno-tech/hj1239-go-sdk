package packet

import (
	"encoding/binary"
	"fmt"

	"github.com/darkinno-tech/hj1239-go-sdk/codec"
	"github.com/darkinno-tech/hj1239-go-sdk/model"
)

// ParsePlatformPacket 从字节流解析企业平台数据包
// 参考 GB1239 Section 5.6.3 Table 16
func ParsePlatformPacket(data []byte) (*model.PlatformPacket, error) {
	if len(data) < model.PlatformPacketHeaderSize+1 {
		return nil, model.ErrShortData("PlatformPacket", model.PlatformPacketHeaderSize+1, len(data))
	}

	p := &model.PlatformPacket{}

	// 起始标识 (0x7e 0x7e)
	if data[0] != 0x7e || data[1] != 0x7e {
		return nil, model.ErrInvalidFrame(fmt.Sprintf("invalid start marker: %02x %02x", data[0], data[1]))
	}
	p.StartMarker = "~~"

	// 命令标识
	p.CommandFlag = data[2]

	// 应答标志
	p.ResponseFlag = data[3]

	// VIN
	vin, err := codec.ReadString(data, 4, 17)
	if err != nil {
		return nil, fmt.Errorf("platform packet: read VIN: %w", err)
	}
	p.VIN = vin

	// 加密方式
	p.EncryptMode = data[21]

	// 数据单元长度
	dataLen, err := codec.ReadUint16(data, 22)
	if err != nil {
		return nil, fmt.Errorf("platform packet: read data length: %w", err)
	}
	p.DataLength = dataLen

	// 数据单元
	if dataLen > 0 {
		p.DataUnit, err = codec.ReadBytes(data, 24, int(dataLen))
		if err != nil {
			return nil, fmt.Errorf("platform packet: read data unit: %w", err)
		}
	}

	// BCC 校验码
	checksumOffset := 24 + int(dataLen)
	if checksumOffset >= len(data) {
		return nil, fmt.Errorf("platform packet: data too short for checksum")
	}
	p.Checksum = data[checksumOffset]

	return p, nil
}

// BuildPlatformPacket 构建企业平台数据包
func BuildPlatformPacket(cmdFlag byte, respFlag byte, vin string, encryptMode byte, dataUnit []byte) (*model.PlatformPacket, error) {
	if len(vin) > 17 {
		return nil, fmt.Errorf("VIN too long: %d > 17", len(vin))
	}

	return &model.PlatformPacket{
		StartMarker:  "~~",
		CommandFlag:  cmdFlag,
		ResponseFlag: respFlag,
		VIN:          vin,
		EncryptMode:  encryptMode,
		DataLength:   uint16(len(dataUnit)),
		DataUnit:     dataUnit,
	}, nil
}

// EncodePlatformPacket 将平台数据包编码为字节流
func EncodePlatformPacket(p *model.PlatformPacket) ([]byte, error) {
	// 计算总大小: 2 + 1 + 1 + 17 + 1 + 2 + len(DataUnit) + 1
	totalSize := model.PlatformPacketHeaderSize + int(p.DataLength) + 1
	buf := make([]byte, totalSize)

	// 起始标识
	buf[0] = 0x7e
	buf[1] = 0x7e

	// 命令标识
	buf[2] = p.CommandFlag

	// 应答标志
	buf[3] = p.ResponseFlag

	// VIN (17 bytes, padded with 0x00)
	codec.WriteString(buf, 4, 17, p.VIN)

	// 加密方式
	buf[21] = p.EncryptMode

	// 数据单元长度
	binary.BigEndian.PutUint16(buf[22:24], p.DataLength)

	// 数据单元
	if p.DataLength > 0 {
		copy(buf[24:24+p.DataLength], p.DataUnit)
	}

	// BCC 校验 (从命令标识到数据单元末尾)
	checksumRange := buf[2 : 24+int(p.DataLength)]
	buf[24+int(p.DataLength)] = codec.BCC(checksumRange)

	return buf, nil
}

// VerifyChecksum 验证数据包的 BCC 校验
func VerifyChecksum(data []byte) bool {
	return CheckChecksum(data) == nil
}

// CheckChecksum verifies the BCC checksum and returns a detailed error on failure.
func CheckChecksum(data []byte) error {
	if len(data) < model.PlatformPacketHeaderSize+1 {
		return model.ErrShortData("checksum", model.PlatformPacketHeaderSize+1, len(data))
	}

	dataLen := binary.BigEndian.Uint16(data[22:24])
	checksumOffset := 24 + int(dataLen)

	if checksumOffset >= len(data) {
		return model.ErrShortData("checksum", checksumOffset+1, len(data))
	}

	checksumRange := data[2:checksumOffset]
	expected := codec.BCC(checksumRange)
	actual := data[checksumOffset]

	if expected != actual {
		return model.WrapError(model.ErrCodeChecksum,
			fmt.Sprintf("expected 0x%02x, got 0x%02x", expected, actual), nil)
	}
	return nil
}

// DecodeVehicleTerminalData 根据命令类型解码车辆终端数据
// Section 4 命令单元 + Section 5 平台数据
func DecodeDataUnit(cmdFlag byte, data []byte) (interface{}, error) {
	switch cmdFlag {
	case 0x01: // 车辆登入
		v := &model.VehicleLoginCode{}
		if err := v.Decode(data); err != nil {
			return nil, err
		}
		return v, nil

	case 0x02: // 实时信息上报
		r := &model.RealTimeInfoReport{}
		if err := r.Decode(data); err != nil {
			return nil, err
		}
		return r, nil

	case 0x03: // 历史信息上报
		h := &model.HistoricalInfoCode{}
		if err := h.Decode(data); err != nil {
			return nil, err
		}
		return h, nil

	case 0x04: // 车辆登出
		v := &model.VehicleLogoutCode{}
		if err := v.Decode(data); err != nil {
			return nil, err
		}
		return v, nil

	case 0x05: // 终端校时
		t := &model.TerminalTimeCalibrationCode{}
		if err := t.Decode(data); err != nil {
			return nil, err
		}
		return t, nil

	case 0x06: // 补充车辆信息
		s := &model.SupplementaryVehicleInfo{}
		if err := s.Decode(data); err != nil {
			return nil, err
		}
		return s, nil

	case 0x07: // 车辆信息 / 企业平台登录
		if len(data) >= model.PlatformLoginCodeSize {
			p := &model.PlatformLoginCode{}
			if err := p.Decode(data); err == nil {
				return p, nil
			}
		}
		// Fallback to vehicle info code
		v := &model.VehicleInfoCode{}
		if err := v.Decode(data); err != nil {
			return nil, fmt.Errorf("failed to decode cmd 0x07: %w", err)
		}
		return v, nil

	case 0x08: // 车辆信息应答 / 企业平台登出
		if len(data) == model.VehicleInfoResponseCodeSize {
			v := &model.VehicleInfoResponseCode{}
			if err := v.Decode(data); err != nil {
				return nil, err
			}
			return v, nil
		}
		p := &model.PlatformLogoutCode{}
		if err := p.Decode(data); err != nil {
			return nil, err
		}
		return p, nil

	case 0x09: // 密钥交换
		k := &model.KeyExchangeCode{}
		if err := k.Decode(data); err != nil {
			return nil, err
		}
		return k, nil

	case 0x0A: // 告警信息上报
		a := &model.AlarmInfoCode{}
		if err := a.Decode(data); err != nil {
			return nil, err
		}
		return a, nil

	case 0x0B: // 文件上传通知
		f := &model.FileUploadNotificationCode{}
		if err := f.Decode(data); err != nil {
			return nil, err
		}
		return f, nil

	case 0x0C: // 文件数据块
		f := &model.FileDataBlockCode{}
		if err := f.Decode(data); err != nil {
			return nil, err
		}
		return f, nil

	case 0x0D: // 文件上传完成
		f := &model.FileUploadCompleteCode{}
		if err := f.Decode(data); err != nil {
			return nil, err
		}
		return f, nil

	case 0x81: // 远程控制请求 (下行)
		c := &model.ControlCode{}
		if err := c.Decode(data); err != nil {
			return nil, err
		}
		return c, nil

	case 0x82: // 远程控制应答 (上行)
		c := &model.ControlResponseCode{}
		if err := c.Decode(data); err != nil {
			return nil, err
		}
		return c, nil
	}

	return data, model.ErrUnknownCmd(cmdFlag)
}
