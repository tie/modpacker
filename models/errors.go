package models

import "errors"

var (
	ErrSumsMismatch     = errors.New("checksum mismatch")
	ErrUnknownModMethod = errors.New("unknown mod method")
	ErrUnknownModAction = errors.New("unknown mod action")
	ErrUnexpectedNode   = errors.New("unexpected html node")
)
