package model

import (
	"encoding/binary"
	"fmt"
	"time"
)

// ============================================================
// Section 4: 车辆终端通讯协议及数据格式
// 参考 GB 17691-2018 附录 Q
// ============================================================

// -------------------------------
// GB 17691 Annex Q 基础消息结构
// -------------------------------

// VehicleLoginCode 车辆登入编码
type VehicleLoginCode struct {
	Time       GB1239Time `hj1239:"offset:0,len:6,desc:数据采集时间"`
	SerialNum  uint16     `hj1239:"offset:6,len:2,desc:登入流水号"`
	ICCID      string     `hj1239:"offset:8,len:20,desc:ICCID"`
	AuthCount  uint8      `hj1239:"offset:28,len:1,desc:认证数据长度"`
	AuthData   []byte     `hj1239:"offset:29,len:var,desc:认证数据"`
	LoginTime  GB1239Time `hj1239:"offset:var,len:6,desc:登入时间"`
}

func (v *VehicleLoginCode) Encode() ([]byte, error) {
	authLen := len(v.AuthData)
	if authLen > 255 {
		authLen = 255
	}
	totalLen := 6 + 2 + 20 + 1 + authLen + 6
	buf := make([]byte, totalLen)

	copy(buf[0:6], v.Time.Bytes())
	binary.BigEndian.PutUint16(buf[6:8], v.SerialNum)
	copy(buf[8:28], padOrTruncate(v.ICCID, 20))
	buf[28] = uint8(authLen)
	copy(buf[29:29+authLen], v.AuthData)
	copy(buf[29+authLen:], v.LoginTime.Bytes())
	return buf, nil
}

func (v *VehicleLoginCode) Decode(b []byte) error {
	if len(b) < 35 {
		return fmt.Errorf("vehicle login code: data too short: %d", len(b))
	}
	var err error
	v.Time, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	v.SerialNum = binary.BigEndian.Uint16(b[6:8])
	v.ICCID = trimString(b[8:28])
	v.AuthCount = b[28]
	authLen := int(v.AuthCount)
	if 29+authLen+6 > len(b) {
		return fmt.Errorf("vehicle login code: auth data overflow")
	}
	v.AuthData = make([]byte, authLen)
	copy(v.AuthData, b[29:29+authLen])
	v.LoginTime, err = ParseGB1239Time(b[29+authLen : 29+authLen+6])
	return err
}

func (v *VehicleLoginCode) Size() int {
	return 6 + 2 + 20 + 1 + len(v.AuthData) + 6
}

// VehicleLogoutCode 车辆登出编码
type VehicleLogoutCode struct {
	Time      GB1239Time `hj1239:"offset:0,len:6,desc:登出时间"`
	SerialNum uint16     `hj1239:"offset:6,len:2,desc:登出流水号"`
}

func (v *VehicleLogoutCode) Encode() ([]byte, error) {
	buf := make([]byte, 8)
	copy(buf[0:6], v.Time.Bytes())
	binary.BigEndian.PutUint16(buf[6:8], v.SerialNum)
	return buf, nil
}

func (v *VehicleLogoutCode) Decode(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("vehicle logout code: data too short: %d", len(b))
	}
	var err error
	v.Time, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	v.SerialNum = binary.BigEndian.Uint16(b[6:8])
	return nil
}

func (v *VehicleLogoutCode) Size() int { return 8 }

// TerminalTimeCalibrationCode 终端校时编码
type TerminalTimeCalibrationCode struct {
	Time GB1239Time `hj1239:"offset:0,len:6,desc:校时时间"`
}

func (t *TerminalTimeCalibrationCode) Encode() ([]byte, error) {
	return t.Time.Bytes(), nil
}

func (t *TerminalTimeCalibrationCode) Decode(b []byte) error {
	var err error
	t.Time, err = ParseGB1239Time(b)
	return err
}

func (t *TerminalTimeCalibrationCode) Size() int { return GB1239TimeLen }

// -------------------------------
// Section 4.5.2: 实时信息上报
// Table 2, Table 3, Table 4
// -------------------------------

// SignatureInfo 签名信息 (Table 3)
type SignatureInfo struct {
	RValueLen uint8  `hj1239:"offset:0,len:1,desc:签名R值长度"`
	RValue    []byte `hj1239:"offset:1,len:var,desc:签名R值"`
	SValue    []byte `hj1239:"offset:var,len:var,desc:签名S值"`
}

func (s *SignatureInfo) Encode() ([]byte, error) {
	rLen := len(s.RValue)
	if rLen > 255 {
		rLen = 255
	}
	sLen := len(s.SValue)
	if sLen > 255 {
		sLen = 255
	}
	buf := make([]byte, 1+rLen+1+sLen)
	buf[0] = uint8(rLen)
	copy(buf[1:1+rLen], s.RValue)
	buf[1+rLen] = uint8(sLen)
	copy(buf[2+rLen:], s.SValue)
	return buf, nil
}

func (s *SignatureInfo) Decode(b []byte) error {
	if len(b) < 3 {
		return fmt.Errorf("signature info: data too short: %d", len(b))
	}
	rLen := int(b[0])
	if 1+rLen+1 > len(b) {
		return fmt.Errorf("signature info: R value overflow")
	}
	s.RValue = make([]byte, rLen)
	copy(s.RValue, b[1:1+rLen])
	sLen := int(b[1+rLen])
	if 1+rLen+1+sLen > len(b) {
		return fmt.Errorf("signature info: S value overflow")
	}
	s.SValue = make([]byte, sLen)
	copy(s.SValue, b[2+rLen:2+rLen+sLen])
	return nil
}

func (s *SignatureInfo) Size() int {
	return 1 + len(s.RValue) + 1 + len(s.SValue)
}

// -------------------------------
// Table 5: DPF/SCR 发动机数据
// -------------------------------

// EngineDataDPFSCR DPF/SCR发动机数据 (Table 5)
// 37 bytes fixed size
type EngineDataDPFSCR struct {
	Speed              ScaledUint16 `hj1239:"offset:0,len:2,scale:0.00390625,unit:km/h,desc:车速"`
	AtmPressure        ScaledUint8  `hj1239:"offset:2,len:1,scale:0.5,unit:kPa,desc:大气压力"`
	EngineTorque       ScaledInt8   `hj1239:"offset:3,len:1,scale:1,offset_val:-125,unit:%,desc:发动机基准扭矩百分比"`
	FrictionTorque     ScaledInt8   `hj1239:"offset:4,len:1,scale:1,offset_val:-125,unit:%,desc:摩擦扭矩百分比"`
	EngineSpeed        ScaledUint16 `hj1239:"offset:5,len:2,scale:0.125,unit:rpm,desc:发动机转速"`
	FuelFlow           ScaledUint16 `hj1239:"offset:7,len:2,scale:0.05,unit:L/h,desc:发动机燃料流量"`
	SCRUpstreamNOx     ScaledInt16  `hj1239:"offset:9,len:2,scale:0.05,offset_val:-200,unit:ppm,desc:SCR上游NOx浓度"`
	SCRDownstreamNOx   ScaledInt16  `hj1239:"offset:11,len:2,scale:0.05,offset_val:-200,unit:ppm,desc:SCR下游NOx浓度"`
	ReagentRate        ScaledUint8  `hj1239:"offset:13,len:1,scale:0.4,unit:%,desc:反应剂剩余率"`
	AirFlow            ScaledUint16 `hj1239:"offset:14,len:2,scale:0.05,unit:kg/h,desc:进气量"`
	SCRUpstreamTemp    ScaledInt16  `hj1239:"offset:16,len:2,scale:0.03125,offset_val:-273,unit:degC,desc:SCR上游温度"`
	SCRDownstreamTemp  ScaledInt16  `hj1239:"offset:18,len:2,scale:0.03125,offset_val:-273,unit:degC,desc:SCR下游温度"`
	DPFDiffPressure    ScaledUint16 `hj1239:"offset:20,len:2,scale:0.1,unit:kPa,desc:DPF压差"`
	CoolantTemp        ScaledInt8   `hj1239:"offset:22,len:1,scale:1,offset_val:-40,unit:degC,desc:发动机冷却液温度"`
	UreaLevel          ScaledUint8  `hj1239:"offset:23,len:1,scale:0.4,unit:%,desc:尿素液位"`
	PositionStatus     uint8        `hj1239:"offset:24,len:1,desc:定位状态"`
	Longitude          ScaledUint32 `hj1239:"offset:25,len:4,scale:0.000001,unit:deg,desc:经度"`
	Latitude           ScaledUint32 `hj1239:"offset:29,len:4,scale:0.000001,unit:deg,desc:纬度"`
	Odometer           ScaledUint32 `hj1239:"offset:33,len:4,scale:0.1,unit:km,desc:累计里程"`
}

const EngineDataDPFSCRSize = 37

func (e *EngineDataDPFSCR) Encode() ([]byte, error) {
	buf := make([]byte, EngineDataDPFSCRSize)
	e.Speed.PutUint16(buf[0:2], true)
	e.AtmPressure.PutUint8(buf[2:3])
	e.EngineTorque.PutInt8(buf[3:4])
	e.FrictionTorque.PutInt8(buf[4:5])
	e.EngineSpeed.PutUint16(buf[5:7], true)
	e.FuelFlow.PutUint16(buf[7:9], true)
	e.SCRUpstreamNOx.PutInt16(buf[9:11], true)
	e.SCRDownstreamNOx.PutInt16(buf[11:13], true)
	e.ReagentRate.PutUint8(buf[13:14])
	e.AirFlow.PutUint16(buf[14:16], true)
	e.SCRUpstreamTemp.PutInt16(buf[16:18], true)
	e.SCRDownstreamTemp.PutInt16(buf[18:20], true)
	e.DPFDiffPressure.PutUint16(buf[20:22], true)
	e.CoolantTemp.PutInt8(buf[22:23])
	e.UreaLevel.PutUint8(buf[23:24])
	buf[24] = e.PositionStatus
	e.Longitude.PutUint32(buf[25:29], true)
	e.Latitude.PutUint32(buf[29:33], true)
	e.Odometer.PutUint32(buf[33:37], true)
	return buf, nil
}

func (e *EngineDataDPFSCR) Decode(b []byte) error {
	if len(b) < EngineDataDPFSCRSize {
		return fmt.Errorf("engine data dpf/scr: data too short: %d", len(b))
	}
	e.Speed.DecodeUint16(b[0:2], true)
	e.AtmPressure.DecodeUint8(b[2:3])
	e.EngineTorque.DecodeInt8(b[3:4])
	e.FrictionTorque.DecodeInt8(b[4:5])
	e.EngineSpeed.DecodeUint16(b[5:7], true)
	e.FuelFlow.DecodeUint16(b[7:9], true)
	e.SCRUpstreamNOx.DecodeInt16(b[9:11], true)
	e.SCRDownstreamNOx.DecodeInt16(b[11:13], true)
	e.ReagentRate.DecodeUint8(b[13:14])
	e.AirFlow.DecodeUint16(b[14:16], true)
	e.SCRUpstreamTemp.DecodeInt16(b[16:18], true)
	e.SCRDownstreamTemp.DecodeInt16(b[18:20], true)
	e.DPFDiffPressure.DecodeUint16(b[20:22], true)
	e.CoolantTemp.DecodeInt8(b[22:23])
	e.UreaLevel.DecodeUint8(b[23:24])
	e.PositionStatus = b[24]
	e.Longitude.DecodeUint32(b[25:29], true)
	e.Latitude.DecodeUint32(b[29:33], true)
	e.Odometer.DecodeUint32(b[33:37], true)
	return nil
}

func (e *EngineDataDPFSCR) Size() int { return EngineDataDPFSCRSize }

// -------------------------------
// Table 6: TWC 发动机数据
// -------------------------------

// EngineDataTWC TWC发动机数据 (Table 6)
// 30 bytes fixed size
type EngineDataTWC struct {
	Speed           ScaledUint16 `hj1239:"offset:0,len:2,scale:0.00390625,unit:km/h,desc:车速"`
	AtmPressure     ScaledUint8  `hj1239:"offset:2,len:1,scale:0.5,unit:kPa,desc:大气压力"`
	EngineTorque    ScaledInt8   `hj1239:"offset:3,len:1,scale:1,offset_val:-125,unit:%,desc:发动机基准扭矩百分比"`
	FrictionTorque  ScaledInt8   `hj1239:"offset:4,len:1,scale:1,offset_val:-125,unit:%,desc:摩擦扭矩百分比"`
	EngineSpeed     ScaledUint16 `hj1239:"offset:5,len:2,scale:0.125,unit:rpm,desc:发动机转速"`
	FuelFlow        ScaledUint16 `hj1239:"offset:7,len:2,scale:0.05,unit:L/h,desc:发动机燃料流量"`
	TWCUpstreamO2   ScaledUint16 `hj1239:"offset:9,len:2,scale:0.0000305,unit:ratio,desc:TWC上游氧浓度"`
	TWCUpstreamVoltage ScaledUint8 `hj1239:"offset:11,len:1,scale:0.01,unit:V,desc:TWC上游氧传感器电压"`
	AirFlow         ScaledUint16 `hj1239:"offset:12,len:2,scale:0.05,unit:kg/h,desc:进气量"`
	CoolantTemp     ScaledInt8   `hj1239:"offset:14,len:1,scale:1,offset_val:-40,unit:degC,desc:发动机冷却液温度"`
	TWCUpstreamTemp ScaledInt16  `hj1239:"offset:15,len:2,scale:0.03125,offset_val:-273,unit:degC,desc:TWC上游温度"`
	PositionStatus  uint8        `hj1239:"offset:17,len:1,desc:定位状态"`
	Longitude       ScaledUint32 `hj1239:"offset:18,len:4,scale:0.000001,unit:deg,desc:经度"`
	Latitude        ScaledUint32 `hj1239:"offset:22,len:4,scale:0.000001,unit:deg,desc:纬度"`
	Odometer        ScaledUint32 `hj1239:"offset:26,len:4,scale:0.1,unit:km,desc:累计里程"`
}

const EngineDataTWCSize = 30

func (e *EngineDataTWC) Encode() ([]byte, error) {
	buf := make([]byte, EngineDataTWCSize)
	e.Speed.PutUint16(buf[0:2], true)
	e.AtmPressure.PutUint8(buf[2:3])
	e.EngineTorque.PutInt8(buf[3:4])
	e.FrictionTorque.PutInt8(buf[4:5])
	e.EngineSpeed.PutUint16(buf[5:7], true)
	e.FuelFlow.PutUint16(buf[7:9], true)
	e.TWCUpstreamO2.PutUint16(buf[9:11], true)
	e.TWCUpstreamVoltage.PutUint8(buf[11:12])
	e.AirFlow.PutUint16(buf[12:14], true)
	e.CoolantTemp.PutInt8(buf[14:15])
	e.TWCUpstreamTemp.PutInt16(buf[15:17], true)
	buf[17] = e.PositionStatus
	e.Longitude.PutUint32(buf[18:22], true)
	e.Latitude.PutUint32(buf[22:26], true)
	e.Odometer.PutUint32(buf[26:30], true)
	return buf, nil
}

func (e *EngineDataTWC) Decode(b []byte) error {
	if len(b) < EngineDataTWCSize {
		return fmt.Errorf("engine data twc: data too short: %d", len(b))
	}
	e.Speed.DecodeUint16(b[0:2], true)
	e.AtmPressure.DecodeUint8(b[2:3])
	e.EngineTorque.DecodeInt8(b[3:4])
	e.FrictionTorque.DecodeInt8(b[4:5])
	e.EngineSpeed.DecodeUint16(b[5:7], true)
	e.FuelFlow.DecodeUint16(b[7:9], true)
	e.TWCUpstreamO2.DecodeUint16(b[9:11], true)
	e.TWCUpstreamVoltage.DecodeUint8(b[11:12])
	e.AirFlow.DecodeUint16(b[12:14], true)
	e.CoolantTemp.DecodeInt8(b[14:15])
	e.TWCUpstreamTemp.DecodeInt16(b[15:17], true)
	e.PositionStatus = b[17]
	e.Longitude.DecodeUint32(b[18:22], true)
	e.Latitude.DecodeUint32(b[22:26], true)
	e.Odometer.DecodeUint32(b[26:30], true)
	return nil
}

func (e *EngineDataTWC) Size() int { return EngineDataTWCSize }

// -------------------------------
// Table 7: TWC+NOx 发动机数据
// -------------------------------

// EngineDataTWCNOx TWC+NOx发动机数据 (Table 7)
// 32 bytes fixed size
type EngineDataTWCNOx struct {
	Speed            ScaledUint16 `hj1239:"offset:0,len:2,scale:0.00390625,unit:km/h,desc:车速"`
	AtmPressure      ScaledUint8  `hj1239:"offset:2,len:1,scale:0.5,unit:kPa,desc:大气压力"`
	EngineTorque     ScaledInt8   `hj1239:"offset:3,len:1,scale:1,offset_val:-125,unit:%,desc:发动机基准扭矩百分比"`
	FrictionTorque   ScaledInt8   `hj1239:"offset:4,len:1,scale:1,offset_val:-125,unit:%,desc:摩擦扭矩百分比"`
	EngineSpeed      ScaledUint16 `hj1239:"offset:5,len:2,scale:0.125,unit:rpm,desc:发动机转速"`
	FuelFlow         ScaledUint16 `hj1239:"offset:7,len:2,scale:0.05,unit:L/h,desc:发动机燃料流量"`
	TWCUpstreamO2    ScaledUint16 `hj1239:"offset:9,len:2,scale:0.0000305,unit:ratio,desc:TWC上游氧浓度"`
	TWCUpstreamVoltage ScaledUint8 `hj1239:"offset:11,len:1,scale:0.01,unit:V,desc:TWC上游氧传感器电压"`
	NOxUpstream      ScaledInt16  `hj1239:"offset:12,len:2,scale:0.05,offset_val:-200,unit:ppm,desc:NOx上游浓度"`
	AirFlow          ScaledUint16 `hj1239:"offset:14,len:2,scale:0.05,unit:kg/h,desc:进气量"`
	CoolantTemp      ScaledInt8   `hj1239:"offset:16,len:1,scale:1,offset_val:-40,unit:degC,desc:发动机冷却液温度"`
	SensorTemp       ScaledInt16  `hj1239:"offset:17,len:2,scale:0.03125,offset_val:-273,unit:degC,desc:TWC温度传感器温度"`
	PositionStatus   uint8        `hj1239:"offset:19,len:1,desc:定位状态"`
	Longitude        ScaledUint32 `hj1239:"offset:20,len:4,scale:0.000001,unit:deg,desc:经度"`
	Latitude         ScaledUint32 `hj1239:"offset:24,len:4,scale:0.000001,unit:deg,desc:纬度"`
	Odometer         ScaledUint32 `hj1239:"offset:28,len:4,scale:0.1,unit:km,desc:累计里程"`
}

const EngineDataTWCNOxSize = 32

func (e *EngineDataTWCNOx) Encode() ([]byte, error) {
	buf := make([]byte, EngineDataTWCNOxSize)
	e.Speed.PutUint16(buf[0:2], true)
	e.AtmPressure.PutUint8(buf[2:3])
	e.EngineTorque.PutInt8(buf[3:4])
	e.FrictionTorque.PutInt8(buf[4:5])
	e.EngineSpeed.PutUint16(buf[5:7], true)
	e.FuelFlow.PutUint16(buf[7:9], true)
	e.TWCUpstreamO2.PutUint16(buf[9:11], true)
	e.TWCUpstreamVoltage.PutUint8(buf[11:12])
	e.NOxUpstream.PutInt16(buf[12:14], true)
	e.AirFlow.PutUint16(buf[14:16], true)
	e.CoolantTemp.PutInt8(buf[16:17])
	e.SensorTemp.PutInt16(buf[17:19], true)
	buf[19] = e.PositionStatus
	e.Longitude.PutUint32(buf[20:24], true)
	e.Latitude.PutUint32(buf[24:28], true)
	e.Odometer.PutUint32(buf[28:32], true)
	return buf, nil
}

func (e *EngineDataTWCNOx) Decode(b []byte) error {
	if len(b) < EngineDataTWCNOxSize {
		return fmt.Errorf("engine data twc+nox: data too short: %d", len(b))
	}
	e.Speed.DecodeUint16(b[0:2], true)
	e.AtmPressure.DecodeUint8(b[2:3])
	e.EngineTorque.DecodeInt8(b[3:4])
	e.FrictionTorque.DecodeInt8(b[4:5])
	e.EngineSpeed.DecodeUint16(b[5:7], true)
	e.FuelFlow.DecodeUint16(b[7:9], true)
	e.TWCUpstreamO2.DecodeUint16(b[9:11], true)
	e.TWCUpstreamVoltage.DecodeUint8(b[11:12])
	e.NOxUpstream.DecodeInt16(b[12:14], true)
	e.AirFlow.DecodeUint16(b[14:16], true)
	e.CoolantTemp.DecodeInt8(b[16:17])
	e.SensorTemp.DecodeInt16(b[17:19], true)
	e.PositionStatus = b[19]
	e.Longitude.DecodeUint32(b[20:24], true)
	e.Latitude.DecodeUint32(b[24:28], true)
	e.Odometer.DecodeUint32(b[28:32], true)
	return nil
}

func (e *EngineDataTWCNOx) Size() int { return EngineDataTWCNOxSize }

// -------------------------------
// Table 8: 混合动力/电动发动机数据
// -------------------------------

// EngineDataHybrid 混合动力/电动发动机数据 (Table 8)
type EngineDataHybrid struct {
	MotorSpeed         ScaledUint16 `hj1239:"offset:0,len:2,scale:1,unit:rpm,desc:电机转速"`
	PowerPercent       ScaledUint8  `hj1239:"offset:2,len:1,scale:1,unit:%,desc:动力输出百分比"`
	BatteryVoltage     ScaledUint16 `hj1239:"offset:3,len:2,scale:0.1,unit:V,desc:电池电压"`
	BatteryCurrent     ScaledInt16  `hj1239:"offset:5,len:2,scale:0.1,offset_val:-1000,unit:A,desc:电池电流"`
	BatterySOC         ScaledUint8  `hj1239:"offset:7,len:1,scale:1,unit:%,desc:电池SOC"`
}

const EngineDataHybridSize = 8

func (e *EngineDataHybrid) Encode() ([]byte, error) {
	buf := make([]byte, EngineDataHybridSize)
	e.MotorSpeed.PutUint16(buf[0:2], true)
	e.PowerPercent.PutUint8(buf[2:3])
	e.BatteryVoltage.PutUint16(buf[3:5], true)
	e.BatteryCurrent.PutInt16(buf[5:7], true)
	e.BatterySOC.PutUint8(buf[7:8])
	return buf, nil
}

func (e *EngineDataHybrid) Decode(b []byte) error {
	if len(b) < EngineDataHybridSize {
		return fmt.Errorf("engine data hybrid: data too short: %d", len(b))
	}
	e.MotorSpeed.DecodeUint16(b[0:2], true)
	e.PowerPercent.DecodeUint8(b[2:3])
	e.BatteryVoltage.DecodeUint16(b[3:5], true)
	e.BatteryCurrent.DecodeInt16(b[5:7], true)
	e.BatterySOC.DecodeUint8(b[7:8])
	return nil
}

func (e *EngineDataHybrid) Size() int { return EngineDataHybridSize }

// -------------------------------
// Table 9: 车辆状态位
// -------------------------------

// VehicleStatus 车辆状态位 (Table 9)
type VehicleStatus struct {
	// 诊断就绪状态 (2 bytes, bit flags)
	MilStatus           uint16 `hj1239:"offset:0,len:2,desc:诊断就绪状态"`
	// 诊断支持状态 (2 bytes, bit flags)
	DiagSupportStatus   uint16 `hj1239:"offset:2,len:2,desc:诊断支持状态"`
	// 诊断完成状态 (2 bytes, bit flags)
	DiagCompleteStatus  uint16 `hj1239:"offset:4,len:2,desc:诊断完成状态"`
	// VIN (17 bytes ASCII)
	VIN string `hj1239:"offset:6,len:17,desc:车辆识别代号(VIN)"`
	// 标定识别号 (variable)
	CalibrationID string `hj1239:"offset:23,len:var,desc:标定识别号"`
	// 标定验证码
	CVN string `hj1239:"offset:var,len:var,desc:标定验证码(CVN)"`
	// IUPR 值
	IUPR string `hj1239:"offset:var,len:36,desc:IUPR值"`
	// 故障码
	DTCCount    uint8  `hj1239:"offset:var,len:1,desc:故障码数量"`
	DTCCodes    []byte `hj1239:"offset:var,len:var,desc:故障码列表"`
}

func (v *VehicleStatus) Encode() ([]byte, error) {
	calID := padOrTruncate(v.CalibrationID, 18)
	cvnStr := padOrTruncate(v.CVN, 18)
	iupr := padOrTruncate(v.IUPR, 36)
	dtcCount := len(v.DTCCodes)
	if dtcCount > 255 {
		dtcCount = 255
	}
	dtcLen := dtcCount * 4

	size := 6 + 17 + 18 + 18 + 36 + 1 + dtcLen
	buf := make([]byte, 0, size)

	// Mil status
	ms := make([]byte, 2)
	binary.BigEndian.PutUint16(ms, v.MilStatus)
	buf = append(buf, ms...)

	// Diag support status
	ds := make([]byte, 2)
	binary.BigEndian.PutUint16(ds, v.DiagSupportStatus)
	buf = append(buf, ds...)

	// Diag complete status
	dc := make([]byte, 2)
	binary.BigEndian.PutUint16(dc, v.DiagCompleteStatus)
	buf = append(buf, dc...)

	buf = append(buf, []byte(padOrTruncate(v.VIN, 17))...)
	buf = append(buf, []byte(calID)...)
	buf = append(buf, []byte(cvnStr)...)
	buf = append(buf, []byte(iupr)...)
	buf = append(buf, uint8(dtcCount))
	buf = append(buf, v.DTCCodes...)

	return buf, nil
}

func (v *VehicleStatus) Decode(b []byte) error {
	if len(b) < 6+17 {
		return fmt.Errorf("vehicle status: data too short: %d", len(b))
	}
	pos := 0
	v.MilStatus = binary.BigEndian.Uint16(b[pos : pos+2])
	pos += 2
	v.DiagSupportStatus = binary.BigEndian.Uint16(b[pos : pos+2])
	pos += 2
	v.DiagCompleteStatus = binary.BigEndian.Uint16(b[pos : pos+2])
	pos += 2
	v.VIN = trimString(b[pos : pos+17])
	pos += 17

	if pos+18 > len(b) {
		return nil
	}
	v.CalibrationID = trimString(b[pos : pos+18])
	pos += 18

	if pos+18 > len(b) {
		return nil
	}
	v.CVN = trimString(b[pos : pos+18])
	pos += 18

	if pos+36 > len(b) {
		return nil
	}
	v.IUPR = trimString(b[pos : pos+36])
	pos += 36

	if pos+1 > len(b) {
		return nil
	}
	v.DTCCount = b[pos]
	pos += 1

	dtcLen := int(v.DTCCount) * 4
	if pos+dtcLen > len(b) {
		return nil
	}
	v.DTCCodes = make([]byte, dtcLen)
	copy(v.DTCCodes, b[pos:pos+dtcLen])

	return nil
}

func (v *VehicleStatus) Size() int {
	dtcLen := len(v.DTCCodes)
	return 6 + 17 + 18 + 18 + 36 + 1 + dtcLen
}

// -------------------------------
// Section 4.5.2.2: 实时信息上报 整体结构
// -------------------------------

// RealTimeInfoBlock 单个实时信息块
type RealTimeInfoBlock struct {
	InfoTypeFlag byte        `hj1239:"offset:0,len:1,desc:信息类型标志"`
	CollectTime  GB1239Time  `hj1239:"offset:1,len:6,desc:信息采集时间"`
	InfoBody     interface{} `hj1239:"offset:7,len:var,desc:信息体"`
	infoBodyRaw  []byte
}

func (r *RealTimeInfoBlock) Encode() ([]byte, error) {
	buf := []byte{r.InfoTypeFlag}
	buf = append(buf, r.CollectTime.Bytes()...)

	if r.infoBodyRaw != nil {
		buf = append(buf, r.infoBodyRaw...)
	}
	return buf, nil
}

func (r *RealTimeInfoBlock) Decode(b []byte) error {
	if len(b) < 7 {
		return fmt.Errorf("realtime info block: data too short: %d", len(b))
	}
	r.InfoTypeFlag = b[0]
	var err error
	r.CollectTime, err = ParseGB1239Time(b[1:7])
	if err != nil {
		return err
	}
	r.infoBodyRaw = make([]byte, len(b)-7)
	copy(r.infoBodyRaw, b[7:])
	return nil
}

func (r *RealTimeInfoBlock) Size() int {
	bodyLen := 0
	if r.infoBodyRaw != nil {
		bodyLen = len(r.infoBodyRaw)
	}
	return 1 + GB1239TimeLen + bodyLen
}

func (r *RealTimeInfoBlock) BodyRaw() []byte {
	return r.infoBodyRaw
}

// RealTimeInfoReport 实时信息上报 (Section 4.5.2)
type RealTimeInfoReport struct {
	DataTime      GB1239Time         `hj1239:"offset:0,len:6,desc:数据发送时间"`
	SerialNum     uint16             `hj1239:"offset:6,len:2,desc:信息流水号"`
	InfoBlocks    []RealTimeInfoBlock `hj1239:"offset:8,len:var,desc:信息块列表"`
	SignatureInfo *SignatureInfo      `hj1239:"offset:var,len:var,desc:签名信息"`
	dataRaw       []byte
}

func (r *RealTimeInfoReport) Encode() ([]byte, error) {
	buf := r.DataTime.Bytes()
	serial := make([]byte, 2)
	binary.BigEndian.PutUint16(serial, r.SerialNum)
	buf = append(buf, serial...)

	for _, block := range r.InfoBlocks {
		blockData, err := block.Encode()
		if err != nil {
			return nil, err
		}
		buf = append(buf, blockData...)
	}

	if r.SignatureInfo != nil {
		sigData, err := r.SignatureInfo.Encode()
		if err != nil {
			return nil, err
		}
		buf = append(buf, sigData...)
	}
	return buf, nil
}

func (r *RealTimeInfoReport) Decode(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("realtime info report: data too short: %d", len(b))
	}
	r.dataRaw = make([]byte, len(b))
	copy(r.dataRaw, b)

	var err error
	r.DataTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	r.SerialNum = binary.BigEndian.Uint16(b[6:8])

	pos := 8
	for pos+7 <= len(b) {
		infoType := b[pos]
		collectTime, err := ParseGB1239Time(b[pos+1 : pos+7])
		if err != nil {
			break
		}
		bodySize := infoBodySize(infoType, b[pos+7:])
		if bodySize < 0 || pos+7+bodySize > len(b) {
			break
		}
		block := RealTimeInfoBlock{
			InfoTypeFlag: infoType,
			CollectTime:  collectTime,
			infoBodyRaw:  make([]byte, bodySize),
		}
		copy(block.infoBodyRaw, b[pos+7:pos+7+bodySize])
		r.InfoBlocks = append(r.InfoBlocks, block)
		pos += 7 + bodySize
	}
	return nil
}

func (r *RealTimeInfoReport) Size() int {
	size := 6 + 2
	for _, block := range r.InfoBlocks {
		size += block.Size()
	}
	if r.SignatureInfo != nil {
		size += r.SignatureInfo.Size()
	}
	return size
}

// -------------------------------
// Section 4.5.5: 车辆信息 (VehicleInfo)
// Table 12
// -------------------------------

type VehicleInfoCode struct {
	CollectTime GB1239Time `hj1239:"offset:0,len:6,desc:数据采集时间"`
	ChipID      string     `hj1239:"offset:6,len:16,desc:芯片ID"`
	Key         string     `hj1239:"offset:22,len:64,desc:密钥"`
	VIN         string     `hj1239:"offset:86,len:17,desc:VIN"`
	Signature   string     `hj1239:"offset:103,len:var,desc:签名信息"`
}

func (v *VehicleInfoCode) Encode() ([]byte, error) {
	buf := v.CollectTime.Bytes()
	buf = append(buf, []byte(padOrTruncate(v.ChipID, 16))...)
	buf = append(buf, []byte(padOrTruncate(v.Key, 64))...)
	buf = append(buf, []byte(padOrTruncate(v.VIN, 17))...)
	buf = append(buf, []byte(v.Signature)...)
	return buf, nil
}

func (v *VehicleInfoCode) Decode(b []byte) error {
	if len(b) < 6+16+64+17 {
		return fmt.Errorf("vehicle info: data too short: %d", len(b))
	}
	var err error
	v.CollectTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	v.ChipID = trimString(b[6:22])
	v.Key = trimString(b[22:86])
	v.VIN = trimString(b[86:103])
	if len(b) > 103 {
		v.Signature = string(b[103:])
	}
	return nil
}

func (v *VehicleInfoCode) Size() int {
	return GB1239TimeLen + 16 + 64 + 17 + len(v.Signature)
}

// VehicleInfoResponseCode 车辆信息应答 (Table 13)
type VehicleInfoResponseCode struct {
	Status  uint8 `hj1239:"offset:0,len:1,desc:状态码 0x01=成功 0x02=失败"`
	Message uint8 `hj1239:"offset:1,len:1,desc:消息 成功=0x00 失败=0x01芯片已激活/0x02VIN错误"`
}

func (v *VehicleInfoResponseCode) Encode() ([]byte, error) {
	return []byte{v.Status, v.Message}, nil
}

func (v *VehicleInfoResponseCode) Decode(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("vehicle info response: data too short: %d", len(b))
	}
	v.Status = b[0]
	v.Message = b[1]
	return nil
}

func (v *VehicleInfoResponseCode) Size() int { return 2 }

const VehicleInfoResponseCodeSize = 2

// -------------------------------
// Section 4.5.4: 车辆终端补充参数 (Annex A)
// Table A.1
// -------------------------------

type TerminalSupplementCode struct {
	CollectTime    GB1239Time  `hj1239:"offset:0,len:6,desc:数据采集时间"`
	SerialNum      uint16      `hj1239:"offset:6,len:2,desc:流水号"`
	PositionStatus uint8       `hj1239:"offset:8,len:1,desc:定位状态"`
	Longitude      ScaledUint32 `hj1239:"offset:9,len:4,scale:0.000001,unit:deg,desc:经度"`
	Latitude       ScaledUint32 `hj1239:"offset:13,len:4,scale:0.000001,unit:deg,desc:纬度"`
}

const TerminalSupplementCodeSize = 17

func (t *TerminalSupplementCode) Encode() ([]byte, error) {
	buf := make([]byte, TerminalSupplementCodeSize)
	copy(buf[0:6], t.CollectTime.Bytes())
	binary.BigEndian.PutUint16(buf[6:8], t.SerialNum)
	buf[8] = t.PositionStatus
	t.Longitude.PutUint32(buf[9:13], true)
	t.Latitude.PutUint32(buf[13:17], true)
	return buf, nil
}

func (t *TerminalSupplementCode) Decode(b []byte) error {
	if len(b) < TerminalSupplementCodeSize {
		return fmt.Errorf("terminal supplement: data too short: %d", len(b))
	}
	var err error
	t.CollectTime, err = ParseGB1239Time(b[0:6])
	if err != nil {
		return err
	}
	t.SerialNum = binary.BigEndian.Uint16(b[6:8])
	t.PositionStatus = b[8]
	t.Longitude.DecodeUint32(b[9:13], true)
	t.Latitude.DecodeUint32(b[13:17], true)
	return nil
}

func (t *TerminalSupplementCode) Size() int { return TerminalSupplementCodeSize }

// -------------------------------
// InfoType 0x01: OBD 信息体
// Table 9: VehicleStatus 衍生
// -------------------------------

// OBDInfoBody OBD 诊断信息体 (用于 RealTimeInfoBlock InfoTypeFlag=0x01)
type OBDInfoBody struct {
	MilStatus          uint16 `hj1239:"offset:0,len:2,desc:诊断就绪状态(MIL)"`
	DiagSupportStatus  uint16 `hj1239:"offset:2,len:2,desc:诊断支持状态"`
	DiagCompleteStatus uint16 `hj1239:"offset:4,len:2,desc:诊断完成状态"`
	VIN                string `hj1239:"offset:6,len:17,desc:车辆识别代号(VIN)"`
	CalibrationID      string `hj1239:"offset:23,len:18,desc:标定识别号"`
	CVN                string `hj1239:"offset:41,len:18,desc:标定验证码(CVN)"`
	IUPR               string `hj1239:"offset:59,len:36,desc:IUPR值"`
	DTCCount           uint8  `hj1239:"offset:95,len:1,desc:故障码数量"`
	DTCCodes           []byte `hj1239:"offset:96,len:var,desc:故障码列表"`
}

func (o *OBDInfoBody) Encode() ([]byte, error) {
	vin := padOrTruncate(o.VIN, 17)
	calID := padOrTruncate(o.CalibrationID, 18)
	cvn := padOrTruncate(o.CVN, 18)
	iupr := padOrTruncate(o.IUPR, 36)

	buf := make([]byte, 0, 96+len(o.DTCCodes))
	b2 := make([]byte, 2)
	binary.BigEndian.PutUint16(b2, o.MilStatus)
	buf = append(buf, b2...)
	binary.BigEndian.PutUint16(b2, o.DiagSupportStatus)
	buf = append(buf, b2...)
	binary.BigEndian.PutUint16(b2, o.DiagCompleteStatus)
	buf = append(buf, b2...)
	buf = append(buf, []byte(vin)...)
	buf = append(buf, []byte(calID)...)
	buf = append(buf, []byte(cvn)...)
	buf = append(buf, []byte(iupr)...)
	buf = append(buf, o.DTCCount)
	buf = append(buf, o.DTCCodes...)
	return buf, nil
}

func (o *OBDInfoBody) Decode(b []byte) error {
	if len(b) < 96 {
		return ErrShortData("OBDInfoBody", 96, len(b))
	}
	o.MilStatus = binary.BigEndian.Uint16(b[0:2])
	o.DiagSupportStatus = binary.BigEndian.Uint16(b[2:4])
	o.DiagCompleteStatus = binary.BigEndian.Uint16(b[4:6])
	o.VIN = trimString(b[6:23])
	o.CalibrationID = trimString(b[23:41])
	o.CVN = trimString(b[41:59])
	o.IUPR = trimString(b[59:95])
	o.DTCCount = b[95]
	if o.DTCCount > 0 {
		dtcLen := int(o.DTCCount) * 4
		if 96+dtcLen <= len(b) {
			o.DTCCodes = make([]byte, dtcLen)
			copy(o.DTCCodes, b[96:96+dtcLen])
		}
	}
	return nil
}

func (o *OBDInfoBody) Size() int {
	return 96 + len(o.DTCCodes)
}

// OBDMilMasks maps bit positions to MIL names for Table 9 / Table 10 compliance.
// Bit 0 → 0x0001 = Catalyst, Bit 1 → 0x0002 = Heated Catalyst, etc.
const (
	MilBitCatalyst          = 0x0001
	MilBitHeatedCatalyst    = 0x0002
	MilBitEvapSystem        = 0x0004
	MilBitSecondaryAir      = 0x0008
	MilBitACRefrigerant     = 0x0010
	MilBitO2Sensor          = 0x0020
	MilBitO2Heater          = 0x0040
	MilBitEGRVVT            = 0x0080
	MilBitColdStart         = 0x0100
	MilBitBoostPressure     = 0x0200
	MilBitDPF               = 0x0400
	MilBitSCRNOxAdsorber    = 0x0800
	MilBitNMHCCatalyst      = 0x1000
	MilBitMisfireSupport    = 0x2000
	MilBitFuelSystemSupport = 0x4000
	MilBitComprehensiveComp = 0x8000
)

func HasMIL(flags uint16, mask uint16) bool { return flags&mask != 0 }

// -------------------------------
// Positioning status helpers (Table 10)
// -------------------------------

type PositionStatus struct {
	Raw uint8
}

const (
	PosBitValid     = 1 << 0 // 0:无效定位, 1:有效定位
	PosBitWestLng   = 1 << 1 // 0:东经, 1:西经 (longitude East/West)
	PosBitNorthLat  = 1 << 2 // 0:南纬, 1:北纬 (latitude North/South)
)

func (p PositionStatus) IsValid() bool    { return p.Raw&PosBitValid != 0 }
func (p PositionStatus) IsWestLng() bool  { return p.Raw&PosBitWestLng != 0 }
func (p PositionStatus) IsNorthLat() bool { return p.Raw&PosBitNorthLat != 0 }

func ParsePositionStatus(raw uint8) PositionStatus { return PositionStatus{Raw: raw} }

// -------------------------------
// Info body parse dispatch for RealTimeInfoBlock
// -------------------------------

// infoBodySize returns the fixed body size for a given info type flag.
// Returns -1 if the type is unknown or the data is too short to determine size.
func infoBodySize(flag byte, data []byte) int {
	switch flag {
	case 0x01: // OBD — variable length, needs header parse
		if len(data) < 96 {
			return -1
		}
		dtcCnt := int(data[95])
		return 96 + dtcCnt*4
	case 0x02: // Engine DPF/SCR
		return EngineDataDPFSCRSize
	case 0x03: // Engine TWC
		return EngineDataTWCSize
	case 0x04: // Engine Hybrid
		return EngineDataHybridSize
	case 0x05: // Engine TWC+NOx
		return EngineDataTWCNOxSize
	case 0x80: // Supplement
		return SupplementaryVehicleInfoSize
	default:
		return -1
	}
}

// ParseInfoBody parses the raw body bytes into a typed struct based on InfoTypeFlag.
func ParseInfoBody(flag byte, raw []byte) (interface{}, error) {
	switch flag {
	case 0x01: // OBD
		v := &OBDInfoBody{}
		if err := v.Decode(raw); err != nil {
			return nil, err
		}
		return v, nil

	case 0x02: // Engine DPF/SCR
		v := &EngineDataDPFSCR{}
		if err := v.Decode(raw); err != nil {
			return nil, err
		}
		return v, nil

	case 0x03: // Engine TWC
		v := &EngineDataTWC{}
		if err := v.Decode(raw); err != nil {
			return nil, err
		}
		return v, nil

	case 0x04: // Engine Hybrid
		v := &EngineDataHybrid{}
		if err := v.Decode(raw); err != nil {
			return nil, err
		}
		return v, nil

	case 0x05: // Engine TWC+NOx
		v := &EngineDataTWCNOx{}
		if err := v.Decode(raw); err != nil {
			return nil, err
		}
		return v, nil

	case 0x80: // Supplement
		v := &SupplementaryVehicleInfo{}
		if err := v.Decode(raw); err != nil {
			return nil, err
		}
		return v, nil
	}
	return raw, fmt.Errorf("unknown info type flag: 0x%02x", flag)
}

// -------------------------------
// Helper constructors
// -------------------------------

// NewVehicleLoginCode creates a VehicleLoginCode for vehicle terminal login.
func NewVehicleLoginCode(iccid string, vin string) *VehicleLoginCode {
	now := NewGB1239Time(Clock())
	return &VehicleLoginCode{
		Time:      now,
		SerialNum: 1,
		ICCID:     iccid,
		AuthData:  []byte(vin),
		LoginTime: now,
	}
}

// NewVehicleLogoutCode creates a VehicleLogoutCode.
func NewVehicleLogoutCode(serial uint16) *VehicleLogoutCode {
	return &VehicleLogoutCode{
		Time:      NewGB1239Time(Clock()),
		SerialNum: serial,
	}
}

// NewPlatformLoginCode creates a platform login data unit.
func NewPlatformLoginCode(username, password string, encryptMode byte) *PlatformLoginCode {
	return &PlatformLoginCode{
		AccessTime:  NewGB1239Time(Clock()),
		SerialNum:   1,
		Username:    username,
		Password:    password,
		EncryptMode: encryptMode,
	}
}

// NewPlatformLogoutCode creates a platform logout data unit.
func NewPlatformLogoutCode(serial uint16) *PlatformLogoutCode {
	return &PlatformLogoutCode{
		LogoutTime: NewGB1239Time(Clock()),
		SerialNum:  serial,
	}
}

// NewEngineDataDPFSCR creates a populated EngineDataDPFSCR from real values.
func NewEngineDataDPFSCR(speed, engineSpeed, atmPressure, fuelFlow float64,
	scrUpNOx, scrDownNOx, scrUpTemp, scrDownTemp float64,
	dpfPressure, coolantTemp, ureaLevel float64,
	engineTorque, frictionTorque, airFlow float64,
	longitude, latitude, odometer float64,
	posValid, isNorth bool) *EngineDataDPFSCR {

	var posStatus uint8
	if posValid {
		posStatus |= PosBitValid
	}
	if isNorth {
		posStatus |= PosBitNorthLat
	}

	return &EngineDataDPFSCR{
		Speed:             NewScaledUint16(speed, 1.0/256.0, 0),
		AtmPressure:       NewScaledUint8(atmPressure, 0.5, 0),
		EngineTorque:      NewScaledInt8(engineTorque, 1, -125),
		FrictionTorque:    NewScaledInt8(frictionTorque, 1, -125),
		EngineSpeed:       NewScaledUint16(engineSpeed, 0.125, 0),
		FuelFlow:          NewScaledUint16(fuelFlow, 0.05, 0),
		SCRUpstreamNOx:    NewScaledInt16(scrUpNOx, 0.05, -200),
		SCRDownstreamNOx:  NewScaledInt16(scrDownNOx, 0.05, -200),
		ReagentRate:       NewScaledUint8(ureaLevel, 0.4, 0),
		AirFlow:           NewScaledUint16(airFlow, 0.05, 0),
		SCRUpstreamTemp:   NewScaledInt16(scrUpTemp, 0.03125, -273),
		SCRDownstreamTemp: NewScaledInt16(scrDownTemp, 0.03125, -273),
		DPFDiffPressure:   NewScaledUint16(dpfPressure, 0.1, 0),
		CoolantTemp:       NewScaledInt8(coolantTemp, 1, -40),
		UreaLevel:         NewScaledUint8(ureaLevel, 0.4, 0),
		PositionStatus:    posStatus,
		Longitude:         NewScaledUint32(longitude, 0.000001, 0),
		Latitude:          NewScaledUint32(latitude, 0.000001, 0),
		Odometer:          NewScaledUint32(odometer, 0.1, 0),
	}
}

// NewEngineDataTWC creates a populated EngineDataTWC from real values.
func NewEngineDataTWC(speed, engineSpeed, atmPressure, fuelFlow float64,
	engineTorque, frictionTorque float64,
	twcO2, twcVoltage, airFlow, coolantTemp, twcTemp float64,
	longitude, latitude, odometer float64,
	posValid, isNorth bool) *EngineDataTWC {

	var posStatus uint8
	if posValid {
		posStatus |= PosBitValid
	}
	if isNorth {
		posStatus |= PosBitNorthLat
	}

	return &EngineDataTWC{
		Speed:              NewScaledUint16(speed, 1.0/256.0, 0),
		AtmPressure:        NewScaledUint8(atmPressure, 0.5, 0),
		EngineTorque:       NewScaledInt8(engineTorque, 1, -125),
		FrictionTorque:     NewScaledInt8(frictionTorque, 1, -125),
		EngineSpeed:        NewScaledUint16(engineSpeed, 0.125, 0),
		FuelFlow:           NewScaledUint16(fuelFlow, 0.05, 0),
		TWCUpstreamO2:      NewScaledUint16(twcO2, 0.0000305, 0),
		TWCUpstreamVoltage: NewScaledUint8(twcVoltage, 0.01, 0),
		AirFlow:            NewScaledUint16(airFlow, 0.05, 0),
		CoolantTemp:        NewScaledInt8(coolantTemp, 1, -40),
		TWCUpstreamTemp:    NewScaledInt16(twcTemp, 0.03125, -273),
		PositionStatus:     posStatus,
		Longitude:          NewScaledUint32(longitude, 0.000001, 0),
		Latitude:           NewScaledUint32(latitude, 0.000001, 0),
		Odometer:           NewScaledUint32(odometer, 0.1, 0),
	}
}

// NewRealTimeInfoBlock creates a RealTimeInfoBlock with the given raw body.
func NewRealTimeInfoBlock(infoType byte, collectTime GB1239Time, bodyRaw []byte) RealTimeInfoBlock {
	return RealTimeInfoBlock{
		InfoTypeFlag: infoType,
		CollectTime:  collectTime,
		infoBodyRaw:  bodyRaw,
	}
}

// Clock is the time source for helper constructors.
// Defaults to time.Now; override for testing.
var Clock = func() time.Time { return time.Now() }

// padOrTruncate 填充或截断字符串到指定长度
func padOrTruncate(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	b := make([]byte, length)
	copy(b, s)
	return string(b)
}

// trimString 去除字符串末尾的空字节
func trimString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
