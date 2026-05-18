package transport_test

import (
	"bytes"
	"testing"

	"github.com/DarkInno/hj1239-go-sdk/transport"
)

func TestEscapeUnescape(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"plain", []byte("hello world")},
		{"with_7E", []byte{0x7E, 0x01, 0x02}},
		{"with_7D", []byte{0x7D, 0x01, 0x02}},
		{"with_both", []byte{0x7E, 0x01, 0x7D, 0x02, 0x7E}},
		{"start_7E7E", []byte{0x7E, 0x7E, 0x01, 0x02}},
		{"only_7E", []byte{0x7E}},
		{"only_7D", []byte{0x7D}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := transport.Escape(tt.data)
			unescaped, err := transport.Unescape(escaped)
			if err != nil {
				t.Fatalf("unescape: %v", err)
			}
			if !bytes.Equal(unescaped, tt.data) {
				t.Errorf("round-trip mismatch:\n  original: %#v\n  escaped:  %#v\n  result:   %#v", tt.data, escaped, unescaped)
			}
		})
	}
}

func TestEscapeExpands(t *testing.T) {
	data := []byte{0x7E, 0x7D, 0x7E}
	escaped := transport.Escape(data)
	if len(escaped) != 6 {
		t.Errorf("expected 6 bytes (3 specials x 2), got %d: %#v", len(escaped), escaped)
	}
}

func TestFrameDeframe(t *testing.T) {
	data := []byte{0x01, 0x02, 0x7E, 0x03, 0x7D, 0x04}
	framed := transport.Frame(data)
	if framed[0] != 0x7E || framed[1] != 0x7E {
		t.Fatal("frame must start with 0x7E 0x7E")
	}
	result, err := transport.Deframe(framed)
	if err != nil {
		t.Fatalf("deframe: %v", err)
	}
	if !bytes.Equal(result, data) {
		t.Errorf("frame-deframe mismatch:\n  original: %#v\n  framed:   %#v\n  result:   %#v", data, framed, result)
	}
}

func TestUnescapeInvalidEscape(t *testing.T) {
	data := []byte{0x7D, 0xFF}
	result, err := transport.Unescape(data)
	if err != nil {
		t.Fatalf("unescape: %v", err)
	}
	if !bytes.Equal(result, data) {
		t.Errorf("unknown escape sequence should pass through: got %#v", result)
	}
}
