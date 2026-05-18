package model

import (
	"encoding/binary"
	"fmt"
)

// ============================================================
// Section 5: 企业平台通讯协议及数据结构
// Section 5.6.3: 数据包结构
// ============================================================

// PlatformPacket 企业平台数据包 (Table 16)
// 基于 TCP/IP 的通讯协议
type PlatformPacket struct {
	StartMarker   string `hj1239:"offset:0,len:2,desc:起始标识 '~~' (0x7e 0x7e)"`
	CommandFlag   byte   `hj1239:"offset:2,len:1,desc:命令标识"`
	ResponseFlag  byte   `hj1239:"offset:3,len:1,desc:应答标志"`
	VIN           string `hj1239:"offset:4,len:17,desc:车辆识别代号(VIN)"`
	EncryptMode   byte   `hj1239:"offset:21,len:1,desc:数据加密方式"`
	DataLength    uint16 `hj1239:"offset:22,len:2,desc:数据单元长度"`
	DataUnit      []byte `hj1239:"offset:24,len:var,desc:数据单元"`
	Checksum      byte   `hj1239:"offset:var,len:1,desc:BCC校验码"`
}

const PlatformPacketHeaderSize = 24

func (p *PlatformPacket) Size() int {
	return PlatformPacketHeaderSize + int(p.DataLength) + 1
}

// IsRequest 判断是否为请求包
func (p *PlatformPacket) IsRequest() bool {
	return p.ResponseFlag == 0xFE
}

// -------------------------------
// Section 5.7.1: 企业平台登录 (Table 19)
// -------------------------------

type PlatformLoginCode struct {
	AccessTime  GB1239Time `hj1239:"offset:0,len:6,desc:接入时间"`
	SerialNum   uint16     `hj1239:"offset:6,len:2,desc:登入流水号"`
	Username    string     `hj1239:"offset:8,len:12,desc:平台登录用户名"`
	Password    string     `hj1239:"offset:20,len:20,desc:平台登录密码"`
	EncryptMode byte       `hj1239:"offset:40,len:1,desc:加密方式"`
}

const PlatformLoginCodeSize = 41

func (p *PlatformLoginCode) Encode() ([]byte, error) {
	buf := make([]byte, PlatformLoginCodeSize)
	copy(buf[0:6], p.AccessTime.Bytes())
	binary.BigEndian.PutUint16(buf[6:8], p.SerialNum)
	copy(buf[8:20], []byte(padOrTruncate(p.Username, 12)))
	copy(buf[20:40], []byte(padOrTruncate(p.Password, 20)))
	buf[40] = p.EncryptMode
	return buf, nil
}

func (p *PlatformLoginCode) Decode(b []byte) error {
	if len(b) < PlatformLoginCodeSize {
		return fmt.Errorf("platform login: data too short: %d", len(b))
	}
	var err error
	p.AccessTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	p.SerialNum = binary.BigEndian.Uint16(b[6:8])
	p.Username = trimString(b[8:20])
	p.Password = trimString(b[20:40])
	p.EncryptMode = b[40]
	return nil
}

func (p *PlatformLoginCode) Size() int { return PlatformLoginCodeSize }

// -------------------------------
// Section 5.7.2: 企业平台登出 (Table 20)
// -------------------------------

type PlatformLogoutCode struct {
	LogoutTime GB1239Time `hj1239:"offset:0,len:6,desc:登出时间"`
	SerialNum  uint16     `hj1239:"offset:6,len:2,desc:登出流水号"`
}

const PlatformLogoutCodeSize = 8

func (p *PlatformLogoutCode) Encode() ([]byte, error) {
	buf := make([]byte, PlatformLogoutCodeSize)
	copy(buf[0:6], p.LogoutTime.Bytes())
	binary.BigEndian.PutUint16(buf[6:8], p.SerialNum)
	return buf, nil
}

func (p *PlatformLogoutCode) Decode(b []byte) error {
	if len(b) < PlatformLogoutCodeSize {
		return fmt.Errorf("platform logout: data too short: %d", len(b))
	}
	var err error
	p.LogoutTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	p.SerialNum = binary.BigEndian.Uint16(b[6:8])
	return nil
}

func (p *PlatformLogoutCode) Size() int { return PlatformLogoutCodeSize }

// -------------------------------
// Section 5.7.3: 密钥交换 (Table 21)
// -------------------------------

type KeyExchangeCode struct {
	KeyType      byte       `hj1239:"offset:0,len:1,desc:密钥类型 0x01=SM2 0x02=SM4 0x03=RSA 0x04=AES128"`
	KeyLength    uint16     `hj1239:"offset:1,len:2,desc:密钥长度(字节数)"`
	KeyData      []byte     `hj1239:"offset:3,len:var,desc:密钥数据"`
	EffectiveTime GB1239Time `hj1239:"offset:var,len:6,desc:生效时间"`
	ExpiryTime   GB1239Time `hj1239:"offset:var,len:6,desc:失效时间"`
}

func (k *KeyExchangeCode) Encode() ([]byte, error) {
	keyLen := len(k.KeyData)
	if keyLen > 65531 {
		keyLen = 65531
	}
	totalSize := 1 + 2 + keyLen + 6 + 6
	buf := make([]byte, totalSize)
	buf[0] = k.KeyType
	binary.BigEndian.PutUint16(buf[1:3], uint16(keyLen))
	copy(buf[3:3+keyLen], k.KeyData)
	copy(buf[3+keyLen:3+keyLen+6], k.EffectiveTime.Bytes())
	copy(buf[3+keyLen+6:], k.ExpiryTime.Bytes())
	return buf, nil
}

func (k *KeyExchangeCode) Decode(b []byte) error {
	if len(b) < 1+2 {
		return fmt.Errorf("key exchange: data too short: %d", len(b))
	}
	k.KeyType = b[0]
	keyLen := int(binary.BigEndian.Uint16(b[1:3]))
	if 3+keyLen+12 > len(b) {
		return fmt.Errorf("key exchange: key data overflow")
	}
	k.KeyData = make([]byte, keyLen)
	copy(k.KeyData, b[3:3+keyLen])
	var err error
	k.EffectiveTime, err = ParseGB1239Time(b[3+keyLen : 3+keyLen+6])
	if err != nil {
		return err
	}
	k.ExpiryTime, err = ParseGB1239Time(b[3+keyLen+6 : 3+keyLen+12])
	return err
}

func (k *KeyExchangeCode) Size() int {
	return 1 + 2 + len(k.KeyData) + 6 + 6
}
