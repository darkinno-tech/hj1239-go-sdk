package model

import (
	"fmt"
	"time"
)

// GB1239Time GB1239 时间格式，6字节 GMT+8
// 格式: [年, 月, 日, 时, 分, 秒]
type GB1239Time struct {
	Year   uint8
	Month  uint8
	Day    uint8
	Hour   uint8
	Minute uint8
	Second uint8
}

const GB1239TimeLen = 6

func NewGB1239Time(t time.Time) GB1239Time {
	loc := time.FixedZone("CST", 8*3600)
	t = t.In(loc)
	return GB1239Time{
		Year:   uint8(t.Year() % 100),
		Month:  uint8(t.Month()),
		Day:    uint8(t.Day()),
		Hour:   uint8(t.Hour()),
		Minute: uint8(t.Minute()),
		Second: uint8(t.Second()),
	}
}

func (t GB1239Time) Bytes() []byte {
	return []byte{t.Year, t.Month, t.Day, t.Hour, t.Minute, t.Second}
}

func ParseGB1239Time(b []byte) (GB1239Time, error) {
	if len(b) < GB1239TimeLen {
		return GB1239Time{}, fmt.Errorf("hj1239 time: expected %d bytes, got %d", GB1239TimeLen, len(b))
	}
	return GB1239Time{
		Year:   b[0],
		Month:  b[1],
		Day:    b[2],
		Hour:   b[3],
		Minute: b[4],
		Second: b[5],
	}, nil
}

func (t GB1239Time) Time() time.Time {
	loc := time.FixedZone("CST", 8*3600)
	return time.Date(2000+int(t.Year), time.Month(t.Month), int(t.Day),
		int(t.Hour), int(t.Minute), int(t.Second), 0, loc)
}

// Tag 定义编码 tag 的键名
const TagKey = "hj1239"

// Tag 支持的选项
const (
	TagOffset    = "offset"     // 字节偏移量
	TagLen       = "len"        // 字节长度
	TagScale     = "scale"      // 缩放因子
	TagOffsetVal = "offset_val" // 偏移值（用于计算实际值 = raw * scale + offset_val）
	TagUnit      = "unit"       // 单位
	TagDesc      = "desc"       // 描述
)

// EncoderDecoder 编解码器接口
type EncoderDecoder interface {
	BinaryEncoder
	BinaryDecoder
}

// BinaryEncoder 编码接口
type BinaryEncoder interface {
	Encode() ([]byte, error)
	Size() int
}

// BinaryDecoder 解码接口
type BinaryDecoder interface {
	Decode([]byte) error
}
