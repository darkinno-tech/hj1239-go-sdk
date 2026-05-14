package model

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Scaled types encode binary values with scale factor and offset.
// actual_value = raw_value * scale + offset_val
// Invalid values use 0xFF... markers as defined in the standard.

// --------------------------------------------------
// ScaledUint8
// --------------------------------------------------

type ScaledUint8 struct {
	Raw   uint8
	Value float64
	Valid bool
}

func NewScaledUint8(val float64, scale, offset float64) ScaledUint8 {
	if math.IsNaN(val) {
		return ScaledUint8{Valid: false}
	}
	raw := EncodeUint8(val, scale, offset)
	return ScaledUint8{Raw: raw, Value: val, Valid: raw != 0xFF}
}

func (s *ScaledUint8) PutUint8(buf []byte) {
	if !s.Valid {
		buf[0] = 0xFF
		return
	}
	buf[0] = s.Raw
}

func (s *ScaledUint8) DecodeUint8(b []byte) {
	s.Raw = b[0]
	s.Valid = s.Raw != 0xFF
}

func (s *ScaledUint8) DecodeUint8Scaled(b []byte, scale, offset float64) {
	s.Raw = b[0]
	s.Valid = s.Raw != 0xFF
	if s.Valid {
		s.Value = float64(s.Raw)*scale + offset
	}
}

func (s *ScaledUint8) Float() float64 { return s.Value }

// --------------------------------------------------
// ScaledInt8
// --------------------------------------------------

type ScaledInt8 struct {
	Raw   uint8
	Value float64
	Valid bool
}

func NewScaledInt8(val float64, scale, offset float64) ScaledInt8 {
	if math.IsNaN(val) {
		return ScaledInt8{Valid: false}
	}
	raw := EncodeInt8(val, scale, offset)
	return ScaledInt8{Raw: raw, Value: val, Valid: raw != 0xFF}
}

func (s *ScaledInt8) PutInt8(buf []byte) {
	if !s.Valid {
		buf[0] = 0xFF
		return
	}
	buf[0] = s.Raw
}

func (s *ScaledInt8) DecodeInt8(b []byte) {
	s.Raw = b[0]
	s.Valid = s.Raw != 0xFF
}

func (s *ScaledInt8) DecodeInt8Scaled(b []byte, scale, offset float64) {
	s.Raw = b[0]
	s.Valid = s.Raw != 0xFF
	if s.Valid {
		s.Value = float64(int8(s.Raw))*scale + offset
	}
}

func (s *ScaledInt8) Float() float64 { return s.Value }

// --------------------------------------------------
// ScaledUint16
// --------------------------------------------------

type ScaledUint16 struct {
	Raw   uint16
	Value float64
	Valid bool
}

func NewScaledUint16(val float64, scale, offset float64) ScaledUint16 {
	if math.IsNaN(val) {
		return ScaledUint16{Valid: false}
	}
	raw := EncodeUint16(val, scale, offset)
	return ScaledUint16{Raw: raw, Value: val, Valid: raw != 0xFFFF}
}

func (s *ScaledUint16) PutUint16(buf []byte, bigEndian bool) {
	if !s.Valid {
		buf[0] = 0xFF
		buf[1] = 0xFF
		return
	}
	if bigEndian {
		binary.BigEndian.PutUint16(buf, s.Raw)
	} else {
		binary.LittleEndian.PutUint16(buf, s.Raw)
	}
}

func (s *ScaledUint16) DecodeUint16(b []byte, bigEndian bool) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint16(b)
	} else {
		s.Raw = binary.LittleEndian.Uint16(b)
	}
	s.Valid = s.Raw != 0xFFFF
}

func (s *ScaledUint16) DecodeUint16Scaled(b []byte, bigEndian bool, scale, offset float64) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint16(b)
	} else {
		s.Raw = binary.LittleEndian.Uint16(b)
	}
	s.Valid = s.Raw != 0xFFFF
	if s.Valid {
		s.Value = float64(s.Raw)*scale + offset
	}
}

func (s *ScaledUint16) Float() float64 { return s.Value }

// --------------------------------------------------
// ScaledInt16
// --------------------------------------------------

type ScaledInt16 struct {
	Raw   uint16
	Value float64
	Valid bool
}

func NewScaledInt16(val float64, scale, offset float64) ScaledInt16 {
	if math.IsNaN(val) {
		return ScaledInt16{Valid: false}
	}
	raw := EncodeInt16(val, scale, offset)
	return ScaledInt16{Raw: raw, Value: val, Valid: raw != 0xFFFF}
}

func (s *ScaledInt16) PutInt16(buf []byte, bigEndian bool) {
	if !s.Valid {
		buf[0] = 0xFF
		buf[1] = 0xFF
		return
	}
	if bigEndian {
		binary.BigEndian.PutUint16(buf, s.Raw)
	} else {
		binary.LittleEndian.PutUint16(buf, s.Raw)
	}
}

func (s *ScaledInt16) DecodeInt16(b []byte, bigEndian bool) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint16(b)
	} else {
		s.Raw = binary.LittleEndian.Uint16(b)
	}
	s.Valid = s.Raw != 0xFFFF
}

func (s *ScaledInt16) DecodeInt16Scaled(b []byte, bigEndian bool, scale, offset float64) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint16(b)
	} else {
		s.Raw = binary.LittleEndian.Uint16(b)
	}
	s.Valid = s.Raw != 0xFFFF
	if s.Valid {
		s.Value = float64(int16(s.Raw))*scale + offset
	}
}

func (s *ScaledInt16) Float() float64 { return s.Value }

// --------------------------------------------------
// ScaledUint32
// --------------------------------------------------

type ScaledUint32 struct {
	Raw   uint32
	Value float64
	Valid bool
}

func NewScaledUint32(val float64, scale, offset float64) ScaledUint32 {
	if math.IsNaN(val) {
		return ScaledUint32{Valid: false}
	}
	raw := EncodeUint32(val, scale, offset)
	return ScaledUint32{Raw: raw, Value: val, Valid: raw != 0xFFFFFFFF}
}

func (s *ScaledUint32) PutUint32(buf []byte, bigEndian bool) {
	if !s.Valid {
		buf[0] = 0xFF
		buf[1] = 0xFF
		buf[2] = 0xFF
		buf[3] = 0xFF
		return
	}
	if bigEndian {
		binary.BigEndian.PutUint32(buf, s.Raw)
	} else {
		binary.LittleEndian.PutUint32(buf, s.Raw)
	}
}

func (s *ScaledUint32) DecodeUint32(b []byte, bigEndian bool) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint32(b)
	} else {
		s.Raw = binary.LittleEndian.Uint32(b)
	}
	s.Valid = s.Raw != 0xFFFFFFFF
}

func (s *ScaledUint32) DecodeUint32Scaled(b []byte, bigEndian bool, scale, offset float64) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint32(b)
	} else {
		s.Raw = binary.LittleEndian.Uint32(b)
	}
	s.Valid = s.Raw != 0xFFFFFFFF
	if s.Valid {
		s.Value = float64(s.Raw)*scale + offset
	}
}

func (s *ScaledUint32) Float() float64 { return s.Value }

// --------------------------------------------------
// ScaledInt32
// --------------------------------------------------

type ScaledInt32 struct {
	Raw   uint32
	Value float64
	Valid bool
}

func NewScaledInt32(val float64, scale, offset float64) ScaledInt32 {
	if math.IsNaN(val) {
		return ScaledInt32{Valid: false}
	}
	v := (val - offset) / scale
	if v < -2147483648 {
		v = -2147483648
	}
	if v > 2147483647 {
		v = 2147483647
	}
	raw := uint32(int32(v))
	return ScaledInt32{Raw: raw, Value: val, Valid: raw != 0xFFFFFFFF}
}

func (s *ScaledInt32) PutInt32(buf []byte, bigEndian bool) {
	if !s.Valid {
		buf[0] = 0xFF
		buf[1] = 0xFF
		buf[2] = 0xFF
		buf[3] = 0xFF
		return
	}
	if bigEndian {
		binary.BigEndian.PutUint32(buf, s.Raw)
	} else {
		binary.LittleEndian.PutUint32(buf, s.Raw)
	}
}

func (s *ScaledInt32) DecodeInt32(b []byte, bigEndian bool) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint32(b)
	} else {
		s.Raw = binary.LittleEndian.Uint32(b)
	}
	s.Valid = s.Raw != 0xFFFFFFFF
}

func (s *ScaledInt32) DecodeInt32Scaled(b []byte, bigEndian bool, scale, offset float64) {
	if bigEndian {
		s.Raw = binary.BigEndian.Uint32(b)
	} else {
		s.Raw = binary.LittleEndian.Uint32(b)
	}
	s.Valid = s.Raw != 0xFFFFFFFF
	if s.Valid {
		s.Value = float64(int32(s.Raw))*scale + offset
	}
}

func (s *ScaledInt32) Float() float64 { return s.Value }

// --------------------------------------------------
// Scale helpers (function form)
// --------------------------------------------------

func ScaleUint8(raw uint8, scale, offset float64) float64 {
	if raw == 0xFF {
		return math.NaN()
	}
	return float64(raw)*scale + offset
}

func ScaleInt8(raw uint8, scale, offset float64) float64 {
	if raw == 0xFF {
		return math.NaN()
	}
	return float64(int8(raw))*scale + offset
}

func ScaleUint16(raw uint16, scale, offset float64) float64 {
	if raw == 0xFFFF {
		return math.NaN()
	}
	return float64(raw)*scale + offset
}

func ScaleInt16(raw uint16, scale, offset float64) float64 {
	if raw == 0xFFFF {
		return math.NaN()
	}
	return float64(int16(raw))*scale + offset
}

func ScaleUint32(raw uint32, scale, offset float64) float64 {
	if raw == 0xFFFFFFFF {
		return math.NaN()
	}
	return float64(raw)*scale + offset
}

// Encode helpers

func EncodeUint8(value, scale, offset float64) uint8 {
	if math.IsNaN(value) {
		return 0xFF
	}
	v := (value - offset) / scale
	if v < 0 {
		v = 0
	}
	if v > 254 {
		v = 254
	}
	return uint8(v)
}

func EncodeInt8(value, scale, offset float64) uint8 {
	if math.IsNaN(value) {
		return 0xFF
	}
	v := (value - offset) / scale
	if v < -128 {
		v = -128
	}
	if v > 127 {
		v = 127
	}
	raw := uint8(int8(v))
	if raw == 0xFF {
		return 0xFE // avoid invalid marker collision
	}
	return raw
}

func EncodeUint16(value, scale, offset float64) uint16 {
	if math.IsNaN(value) {
		return 0xFFFF
	}
	v := (value - offset) / scale
	if v < 0 {
		v = 0
	}
	if v > 65534 {
		v = 65534
	}
	return uint16(v)
}

func EncodeInt16(value, scale, offset float64) uint16 {
	if math.IsNaN(value) {
		return 0xFFFF
	}
	v := (value - offset) / scale
	if v < -32768 {
		v = -32768
	}
	if v > 32767 {
		v = 32767
	}
	raw := uint16(int16(v))
	if raw == 0xFFFF {
		return 0xFFFE // avoid invalid marker collision
	}
	return raw
}

func EncodeUint32(value, scale, offset float64) uint32 {
	if math.IsNaN(value) {
		return 0xFFFFFFFF
	}
	v := (value - offset) / scale
	if v < 0 {
		v = 0
	}
	if v > 4294967294 {
		v = 4294967294
	}
	return uint32(v)
}

// EncodeWithError is like EncodeUint16 but returns an error when value overflows
func EncodeUint16WithError(value, scale, offset float64) (uint16, error) {
	if math.IsNaN(value) {
		return 0xFFFF, fmt.Errorf("NaN value")
	}
	v := (value - offset) / scale
	if v < 0 {
		return 0, fmt.Errorf("value %f underflows min 0 (raw %.0f)", value, v)
	}
	if v > 65534 {
		return 65534, fmt.Errorf("value %f overflows max 65534 (raw %.0f)", value, v)
	}
	return uint16(v), nil
}
