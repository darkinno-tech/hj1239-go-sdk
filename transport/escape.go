package transport

// Escape applies HJ 1239.3 TCP frame escaping.
// Replaces 0x7E with 0x7D 0x02 and 0x7D with 0x7D 0x01.
func Escape(data []byte) []byte {
	var buf []byte
	for _, b := range data {
		switch b {
		case 0x7E:
			buf = append(buf, 0x7D, 0x02)
		case 0x7D:
			buf = append(buf, 0x7D, 0x01)
		default:
			buf = append(buf, b)
		}
	}
	return buf
}

// Unescape reverses HJ 1239.3 TCP frame escaping.
// Restores 0x7D 0x02 to 0x7E and 0x7D 0x01 to 0x7D.
func Unescape(data []byte) ([]byte, error) {
	var buf []byte
	for i := 0; i < len(data); i++ {
		if data[i] == 0x7D && i+1 < len(data) {
			switch data[i+1] {
			case 0x02:
				buf = append(buf, 0x7E)
				i++
				continue
			case 0x01:
				buf = append(buf, 0x7D)
				i++
				continue
			}
		}
		buf = append(buf, data[i])
	}
	return buf, nil
}

// Frame wraps data with start marker 0x7E 0x7E and applies escaping.
func Frame(data []byte) []byte {
	escaped := Escape(data)
	framed := make([]byte, 2+len(escaped))
	framed[0] = 0x7E
	framed[1] = 0x7E
	copy(framed[2:], escaped)
	return framed
}

// Deframe removes start marker and unescapes data.
func Deframe(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0x7E || data[1] != 0x7E {
		return data, nil
	}
	return Unescape(data[2:])
}
