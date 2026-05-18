package model

import (
	"encoding/binary"
	"fmt"
)

// HistoricalInfoBlock 历史信息数据块
// 每个数据块包含一段连续时间内采集的二进制数据
type HistoricalInfoBlock struct {
	BeginTime GB1239Time `hj1239:"offset:0,len:6,desc:开始时间"`
	EndTime   GB1239Time `hj1239:"offset:6,len:6,desc:结束时间"`
	Data      []byte     `hj1239:"offset:12,len:var,desc:历史数据"`
}

func (h *HistoricalInfoBlock) Encode() ([]byte, error) {
	dataLen := len(h.Data)
	if dataLen > 65535 {
		return nil, fmt.Errorf("historical info block: data too long: %d > 65535", dataLen)
	}
	totalSize := 6 + 6 + 2 + dataLen
	buf := make([]byte, totalSize)
	copy(buf[0:6], h.BeginTime.Bytes())
	copy(buf[6:12], h.EndTime.Bytes())
	binary.BigEndian.PutUint16(buf[12:14], uint16(dataLen))
	copy(buf[14:], h.Data)
	return buf, nil
}

func (h *HistoricalInfoBlock) Decode(b []byte) error {
	if len(b) < 14 {
		return fmt.Errorf("historical info block: data too short: %d", len(b))
	}
	var err error
	h.BeginTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	h.EndTime, err = ParseGB1239Time(b[6:12])
	if err != nil {
		return err
	}
	dataLen := int(binary.BigEndian.Uint16(b[12:14]))
	if 14+dataLen > len(b) {
		return fmt.Errorf("historical info block: data overflow")
	}
	h.Data = make([]byte, dataLen)
	copy(h.Data, b[14:14+dataLen])
	return nil
}

func (h *HistoricalInfoBlock) Size() int {
	return 6 + 6 + 2 + len(h.Data)
}

// HistoricalInfoCode 历史信息上报 (Section 4.5.3, cmd 0x03)
type HistoricalInfoCode struct {
	DataTime  GB1239Time            `hj1239:"offset:0,len:6,desc:数据发送时间"`
	SerialNum uint16                `hj1239:"offset:6,len:2,desc:信息流水号"`
	Blocks    []HistoricalInfoBlock `hj1239:"offset:8,len:var,desc:历史数据块列表"`
}

func (h *HistoricalInfoCode) Encode() ([]byte, error) {
	buf := h.DataTime.Bytes()
	serial := make([]byte, 2)
	binary.BigEndian.PutUint16(serial, h.SerialNum)
	buf = append(buf, serial...)

	for _, block := range h.Blocks {
		blockData, err := block.Encode()
		if err != nil {
			return nil, err
		}
		buf = append(buf, blockData...)
	}
	return buf, nil
}

func (h *HistoricalInfoCode) Decode(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("historical info code: data too short: %d", len(b))
	}
	var err error
	h.DataTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	h.SerialNum = binary.BigEndian.Uint16(b[6:8])

	pos := 8
	for pos+14 <= len(b) {
		var block HistoricalInfoBlock
		if err := block.Decode(b[pos:]); err != nil {
			break
		}
		h.Blocks = append(h.Blocks, block)
		pos += block.Size()
	}
	return nil
}

func (h *HistoricalInfoCode) Size() int {
	size := 6 + 2
	for _, block := range h.Blocks {
		size += block.Size()
	}
	return size
}
