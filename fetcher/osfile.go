package fetcher

import (
	"os"

	"gopkg.in/src-d/go-billy.v4"
)

var _ billy.File = osFile{}

type osFile struct {
	*os.File
}

func (osFile) Lock() error {
	return billy.ErrNotSupported
}

func (osFile) Unlock() error {
	return billy.ErrNotSupported
}
