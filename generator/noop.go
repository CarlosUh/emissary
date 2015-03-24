package generator

import (
	"io"
)

type noop struct{}

func (n *noop) Generate(w io.Writer) error {
	return nil
}

var Noop = &noop{}
