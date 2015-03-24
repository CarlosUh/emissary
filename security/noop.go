package security

import (
	"io"
)

type noop struct{}

func (n *noop) Secure(r io.Reader, w io.Writer) error {
	return nil
}

var Noop = &noop{}
