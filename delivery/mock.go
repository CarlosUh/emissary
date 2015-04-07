package delivery

import (
	"io"
	"io/ioutil"
)

type Mock struct {
	Data []byte
}

func (m *Mock) Deliver(r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	m.Data = data

	return nil
}
