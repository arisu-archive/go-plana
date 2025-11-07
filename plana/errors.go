package plana

import (
	"fmt"

	"github.com/arisu-archive/plana-protos/protos"
)

type ErrInvalidSession struct {
	msg   string
	code  protos.WebAPIErrorCode
	inner *ErrWebAPIError
}

func (e ErrInvalidSession) Error() string {
	return e.msg
}

func (e ErrInvalidSession) Code() protos.WebAPIErrorCode {
	return e.code
}

func (e ErrInvalidSession) Unwrap() error {
	return e.inner
}

func NewInvalidSessionError(message string, inner *ErrWebAPIError) *ErrInvalidSession {
	return &ErrInvalidSession{
		msg:   message,
		code:  inner.Code(),
		inner: inner,
	}
}

type ErrWebAPIError struct {
	Packet *protos.ErrorPacket
}

func (e *ErrWebAPIError) Error() string {
	return fmt.Sprintf("web api error occurred code=%d, reason=%s", e.Packet.ErrorCode, e.Packet.Reason)
}

func (e *ErrWebAPIError) Code() protos.WebAPIErrorCode {
	return e.Packet.ErrorCode
}

func NewWebAPIError(packet *protos.ErrorPacket) *ErrWebAPIError {
	return &ErrWebAPIError{
		Packet: packet,
	}
}
