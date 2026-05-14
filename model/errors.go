package model

import "fmt"

// ErrorCode is a machine-readable error code.
type ErrorCode string

const (
	ErrCodeShortData    ErrorCode = "SHORT_DATA"
	ErrCodeInvalidFrame ErrorCode = "INVALID_FRAME"
	ErrCodeChecksum     ErrorCode = "CHECKSUM"
	ErrCodeUnknownCmd   ErrorCode = "UNKNOWN_CMD"
	ErrCodeEncode       ErrorCode = "ENCODE_ERROR"
	ErrCodeDecode       ErrorCode = "DECODE_ERROR"
	ErrCodeOverflow     ErrorCode = "OVERFLOW"
)

// ProtocolError is the standard error type for the GB1239 SDK.
type ProtocolError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *ProtocolError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	return msg
}

func (e *ProtocolError) Unwrap() error { return e.Cause }

func NewError(code ErrorCode, msg string) *ProtocolError {
	return &ProtocolError{Code: code, Message: msg}
}

func WrapError(code ErrorCode, msg string, cause error) *ProtocolError {
	return &ProtocolError{Code: code, Message: msg, Cause: cause}
}

func ErrShortData(field string, expected, actual int) *ProtocolError {
	return &ProtocolError{
		Code:    ErrCodeShortData,
		Message: fmt.Sprintf("field %q: need %d bytes, got %d", field, expected, actual),
	}
}

func ErrInvalidFrame(msg string) *ProtocolError {
	return &ProtocolError{Code: ErrCodeInvalidFrame, Message: msg}
}

func ErrChecksum() *ProtocolError {
	return &ProtocolError{Code: ErrCodeChecksum, Message: "BCC checksum mismatch"}
}

func ErrUnknownCmd(cmd byte) *ProtocolError {
	return &ProtocolError{
		Code:    ErrCodeUnknownCmd,
		Message: fmt.Sprintf("unknown command: 0x%02x", cmd),
	}
}
