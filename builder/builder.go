package builder

import (
	"github.com/tie/modpacker/models"
)

type Builder interface {
	Add(m models.Mod) error
	Close() error
}
