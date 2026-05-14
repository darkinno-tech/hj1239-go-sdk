package model

import (
	"encoding/binary"
	"fmt"
)

// ============================================================
// Annex B: 补充车辆信息数据格式 (Table B.1)
// ============================================================

// SupplementaryVehicleInfo 补充车辆信息 (Annex B)
// 18 bytes fixed size (excluding variable fields)
type SupplementaryVehicleInfo struct {
	TorqueMode         ScaledUint8  `hj1239:"offset:0,len:1,desc:发动机扭矩模式:0=无效/1=转速/2=扭矩/3=转速扭矩/9=保留"`
	AccelPedal         ScaledUint8  `hj1239:"offset:1,len:1,scale:0.4,unit:%,desc:油门踏板开度"`
	CumulativeFuel     ScaledUint32 `hj1239:"offset:2,len:4,scale:0.5,unit:L,desc:累计油耗(发动机油耗)"`
	EngineOilTemp      ScaledInt8   `hj1239:"offset:6,len:1,scale:1,offset_val:-40,unit:degC,desc:发动机机油温度"`
	ActualFuelRate     ScaledUint32 `hj1239:"offset:7,len:4,scale:0.01,unit:ml/h,desc:实际发动机燃料率"`
	CumulativeFuelUsed ScaledUint32 `hj1239:"offset:11,len:4,scale:1,unit:g,desc:累计燃料使用量(发动机燃料使用量)"`
	DPFUpstreamTemp    ScaledInt16  `hj1239:"offset:15,len:2,scale:0.03125,offset_val:-273,unit:degC,desc:DPF上游温度"`
}

const SupplementaryVehicleInfoSize = 17

func (s *SupplementaryVehicleInfo) Encode() ([]byte, error) {
	buf := make([]byte, SupplementaryVehicleInfoSize)
	s.TorqueMode.PutUint8(buf[0:1])
	s.AccelPedal.PutUint8(buf[1:2])
	s.CumulativeFuel.PutUint32(buf[2:6], true)
	s.EngineOilTemp.PutInt8(buf[6:7])
	s.ActualFuelRate.PutUint32(buf[7:11], true)
	s.CumulativeFuelUsed.PutUint32(buf[11:15], true)
	s.DPFUpstreamTemp.PutInt16(buf[15:17], true)
	return buf, nil
}

func (s *SupplementaryVehicleInfo) Decode(b []byte) error {
	if len(b) < SupplementaryVehicleInfoSize {
		return fmt.Errorf("supplementary vehicle info: data too short: %d", len(b))
	}
	s.TorqueMode.DecodeUint8(b[0:1])
	s.AccelPedal.DecodeUint8(b[1:2])
	s.CumulativeFuel.DecodeUint32(b[2:6], true)
	s.EngineOilTemp.DecodeInt8(b[6:7])
	s.ActualFuelRate.DecodeUint32(b[7:11], true)
	s.CumulativeFuelUsed.DecodeUint32(b[11:15], true)
	s.DPFUpstreamTemp.DecodeInt16(b[15:17], true)
	return nil
}

func (s *SupplementaryVehicleInfo) Size() int { return SupplementaryVehicleInfoSize }

// Annex B补充车辆信息 Table B.1 full structure with dynamic fields
// (includes variable-length fields at end)

type AnnexBData struct {
	TorqueMode         uint8  `hj1239:"offset:0,len:1,desc:发动机扭矩模式"`
	AccelPedalRaw      uint8  `hj1239:"offset:1,len:1,desc:油门踏板开度(raw)"`
	CumulativeFuelRaw  uint32 `hj1239:"offset:2,len:4,desc:累计油耗(raw)"`
	EngineOilTempRaw   uint8  `hj1239:"offset:6,len:1,desc:发动机机油温度(raw)"`
	ActualFuelRateRaw  uint32 `hj1239:"offset:7,len:4,desc:实际燃料率(raw)"`
	CumFuelUsedRaw     uint32 `hj1239:"offset:11,len:4,desc:累计燃料使用量(raw)"`
	DPFUpTempRaw       uint16 `hj1239:"offset:15,len:2,desc:DPF上游温度(raw)"`
}

const AnnexBDataFixedSize = 17

func DecodeAnnexBData(b []byte) (*AnnexBData, error) {
	if len(b) < AnnexBDataFixedSize {
		return nil, fmt.Errorf("annex B data: data too short: %d", len(b))
	}
	d := &AnnexBData{}
	d.TorqueMode = b[0]
	d.AccelPedalRaw = b[1]
	d.CumulativeFuelRaw = binary.BigEndian.Uint32(b[2:6])
	d.EngineOilTempRaw = b[6]
	d.ActualFuelRateRaw = binary.BigEndian.Uint32(b[7:11])
	d.CumFuelUsedRaw = binary.BigEndian.Uint32(b[11:15])
	d.DPFUpTempRaw = binary.BigEndian.Uint16(b[15:17])
	return d, nil
}

func (d *AnnexBData) AccelPedal() float64 {
	return ScaleUint8(d.AccelPedalRaw, 0.4, 0)
}

func (d *AnnexBData) CumulativeFuel() float64 {
	return ScaleUint32(d.CumulativeFuelRaw, 0.5, 0)
}

func (d *AnnexBData) EngineOilTemp() float64 {
	return ScaleInt8(d.EngineOilTempRaw, 1, -40)
}

func (d *AnnexBData) ActualFuelRate() float64 {
	return ScaleUint32(d.ActualFuelRateRaw, 0.01, 0)
}

func (d *AnnexBData) CumFuelUsed() float64 {
	return ScaleUint32(d.CumFuelUsedRaw, 1, 0)
}

func (d *AnnexBData) DPFUpstreamTemp() float64 {
	return ScaleInt16(d.DPFUpTempRaw, 0.03125, -273)
}
