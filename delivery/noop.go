package delivery

import (
	"io"
)

type noop struct{}

func (n *noop) Deliver(r io.Reader) error {
	return nil
}

var Noop = &noop{}
