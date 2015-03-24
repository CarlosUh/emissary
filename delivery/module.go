package delivery

import (
	"io"
)

type Module interface {
	Deliver(io.Reader) error
}
