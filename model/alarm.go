package model

import (
	"encoding/binary"
	"fmt"
)

// AlarmInfoCode 告警信息上报
type AlarmInfoCode struct {
	DataTime   GB1239Time `hj1239:"offset:0,len:6,desc:数据发送时间"`
	SerialNum  uint16     `hj1239:"offset:6,len:2,desc:信息流水号"`
	AlarmCount uint8      `hj1239:"offset:8,len:1,desc:告警数量"`
	Alarms     []Alarm    `hj1239:"offset:9,len:var,desc:告警列表"`
}

type Alarm struct {
	AlarmType uint8      `hj1239:"offset:0,len:1,desc:告警类型"`
	AlarmTime GB1239Time `hj1239:"offset:1,len:6,desc:告警时间"`
	AlarmData []byte     `hj1239:"offset:7,len:var,desc:告警数据"`
}

const (
	AlarmTypeOBDFault     uint8 = 0x01
	AlarmTypeEngineData   uint8 = 0x02
	AlarmTypePosition     uint8 = 0x03
	AlarmTypeOverspeed    uint8 = 0x04
	AlarmTypePowerOff     uint8 = 0x05
	AlarmTypeTamper       uint8 = 0x06
	AlarmTypeOther        uint8 = 0xFF
)

func (a *Alarm) Encode() ([]byte, error) {
	dataLen := len(a.AlarmData)
	if dataLen > 65535 {
		return nil, fmt.Errorf("alarm: data too long: %d", dataLen)
	}
	totalSize := 1 + GB1239TimeLen + 2 + dataLen
	buf := make([]byte, totalSize)
	buf[0] = a.AlarmType
	copy(buf[1:7], a.AlarmTime.Bytes())
	binary.BigEndian.PutUint16(buf[7:9], uint16(dataLen))
	copy(buf[9:], a.AlarmData)
	return buf, nil
}

func (a *Alarm) Decode(b []byte) error {
	if len(b) < 9 {
		return fmt.Errorf("alarm: data too short: %d", len(b))
	}
	a.AlarmType = b[0]
	var err error
	a.AlarmTime, err = ParseGB1239Time(b[1:7])
	if err != nil {
		return err
	}
	dataLen := int(binary.BigEndian.Uint16(b[7:9]))
	if 9+dataLen > len(b) {
		return fmt.Errorf("alarm: data overflow")
	}
	a.AlarmData = make([]byte, dataLen)
	copy(a.AlarmData, b[9:9+dataLen])
	return nil
}

func (a *Alarm) Size() int { return 1 + GB1239TimeLen + 2 + len(a.AlarmData) }

func (ac *AlarmInfoCode) Encode() ([]byte, error) {
	buf := ac.DataTime.Bytes()
	serial := make([]byte, 2)
	binary.BigEndian.PutUint16(serial, ac.SerialNum)
	buf = append(buf, serial...)
	buf = append(buf, byte(len(ac.Alarms)))
	for _, alarm := range ac.Alarms {
		alarmData, err := alarm.Encode()
		if err != nil {
			return nil, err
		}
		buf = append(buf, alarmData...)
	}
	return buf, nil
}

func (ac *AlarmInfoCode) Decode(b []byte) error {
	if len(b) < 9 {
		return fmt.Errorf("alarm info code: data too short: %d", len(b))
	}
	var err error
	ac.DataTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	ac.SerialNum = binary.BigEndian.Uint16(b[6:8])
	ac.AlarmCount = b[8]

	pos := 9
	for i := uint8(0); i < ac.AlarmCount && pos < len(b); i++ {
		var alarm Alarm
		if err := alarm.Decode(b[pos:]); err != nil {
			break
		}
		ac.Alarms = append(ac.Alarms, alarm)
		pos += alarm.Size()
	}
	return nil
}

func (ac *AlarmInfoCode) Size() int {
	size := GB1239TimeLen + 2 + 1
	for _, alarm := range ac.Alarms {
		size += alarm.Size()
	}
	return size
}
