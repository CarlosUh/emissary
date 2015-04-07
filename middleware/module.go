package middleware

import (
	"io"
)

type Module interface {
	Passthru(io.Reader, io.Writer) error
}
