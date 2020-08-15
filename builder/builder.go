package builder

import (
	"errors"

	"github.com/tie/modpacker/modpacker"
)

var ErrUnknownModAction = errors.New("unknown mod action")

type Builder interface {
	Add(m modpacker.Mod) error
	Close() error
}
