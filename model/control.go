package model

import (
	"encoding/binary"
	"fmt"
)

// ControlCode 远程控制指令 (平台 → 终端下行)
type ControlCode struct {
	ControlType uint8      `hj1239:"offset:0,len:1,desc:控制类型"`
	SerialNum   uint16     `hj1239:"offset:1,len:2,desc:流水号"`
	ParamCount  uint8      `hj1239:"offset:3,len:1,desc:参数数量"`
	Params      []ControlParam `hj1239:"offset:4,len:var,desc:参数列表"`
}

type ControlParam struct {
	ParamID    uint16 `hj1239:"offset:0,len:2,desc:参数ID"`
	ParamLen   uint16 `hj1239:"offset:2,len:2,desc:参数长度"`
	ParamValue []byte `hj1239:"offset:4,len:var,desc:参数值"`
}

const (
	ControlTypeQuery   uint8 = 0x01
	ControlTypeSet     uint8 = 0x02
	ControlTypeUpgrade uint8 = 0x03
	ControlTypeRestart uint8 = 0x04
)

func (cp *ControlParam) Encode() ([]byte, error) {
	totalSize := 2 + 2 + len(cp.ParamValue)
	buf := make([]byte, totalSize)
	binary.BigEndian.PutUint16(buf[0:2], cp.ParamID)
	binary.BigEndian.PutUint16(buf[2:4], cp.ParamLen)
	copy(buf[4:], cp.ParamValue)
	return buf, nil
}

func (cp *ControlParam) Decode(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("control param: data too short: %d", len(b))
	}
	cp.ParamID = binary.BigEndian.Uint16(b[0:2])
	cp.ParamLen = binary.BigEndian.Uint16(b[2:4])
	if 4+int(cp.ParamLen) > len(b) {
		return fmt.Errorf("control param: data overflow")
	}
	cp.ParamValue = make([]byte, cp.ParamLen)
	copy(cp.ParamValue, b[4:4+cp.ParamLen])
	return nil
}

func (cp *ControlParam) Size() int { return 2 + 2 + int(cp.ParamLen) }

func (c *ControlCode) Encode() ([]byte, error) {
	buf := []byte{c.ControlType}
	serial := make([]byte, 2)
	binary.BigEndian.PutUint16(serial, c.SerialNum)
	buf = append(buf, serial...)
	buf = append(buf, byte(len(c.Params)))
	for _, p := range c.Params {
		paramData, err := p.Encode()
		if err != nil {
			return nil, err
		}
		buf = append(buf, paramData...)
	}
	return buf, nil
}

func (c *ControlCode) Decode(b []byte) error {
	if len(b) < 4 {
		return fmt.Errorf("control code: data too short: %d", len(b))
	}
	c.ControlType = b[0]
	c.SerialNum = binary.BigEndian.Uint16(b[1:3])
	c.ParamCount = b[3]

	pos := 4
	for i := uint8(0); i < c.ParamCount && pos < len(b); i++ {
		var p ControlParam
		if err := p.Decode(b[pos:]); err != nil {
			break
		}
		c.Params = append(c.Params, p)
		pos += p.Size()
	}
	return nil
}

func (c *ControlCode) Size() int {
	size := 1 + 2 + 1
	for _, p := range c.Params {
		size += p.Size()
	}
	return size
}

// ControlResponseCode 远程控制应答 (终端 → 平台上行)
type ControlResponseCode struct {
	SerialNum   uint16     `hj1239:"offset:0,len:2,desc:应答流水号"`
	ControlType uint8      `hj1239:"offset:2,len:1,desc:控制类型"`
	Result      uint8      `hj1239:"offset:3,len:1,desc:执行结果 0x01=成功 0x02=失败"`
	DataLen     uint16     `hj1239:"offset:4,len:2,desc:应答数据长度"`
	Data        []byte     `hj1239:"offset:6,len:var,desc:应答数据"`
}

const ControlResponseCodeHeaderSize = 6

func (cr *ControlResponseCode) Encode() ([]byte, error) {
	totalSize := ControlResponseCodeHeaderSize + len(cr.Data)
	buf := make([]byte, totalSize)
	binary.BigEndian.PutUint16(buf[0:2], cr.SerialNum)
	buf[2] = cr.ControlType
	buf[3] = cr.Result
	binary.BigEndian.PutUint16(buf[4:6], uint16(len(cr.Data)))
	copy(buf[6:], cr.Data)
	return buf, nil
}

func (cr *ControlResponseCode) Decode(b []byte) error {
	if len(b) < ControlResponseCodeHeaderSize {
		return fmt.Errorf("control response: data too short: %d", len(b))
	}
	cr.SerialNum = binary.BigEndian.Uint16(b[0:2])
	cr.ControlType = b[2]
	cr.Result = b[3]
	cr.DataLen = binary.BigEndian.Uint16(b[4:6])
	if 6+int(cr.DataLen) > len(b) {
		return fmt.Errorf("control response: data overflow")
	}
	cr.Data = make([]byte, cr.DataLen)
	copy(cr.Data, b[6:6+cr.DataLen])
	return nil
}

func (cr *ControlResponseCode) Size() int { return ControlResponseCodeHeaderSize + len(cr.Data) }
