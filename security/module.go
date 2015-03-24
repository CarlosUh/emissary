package security

import (
	"io"
)

type Module interface {
	Secure(io.Reader, io.Writer) error
}
