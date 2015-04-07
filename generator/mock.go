package generator

import (
	"io"
)

type Mock struct {
	Data string
}

func (m *Mock) Generate(w io.Writer) error {
	_, err := w.Write([]byte(m.Data))
	return err
}
