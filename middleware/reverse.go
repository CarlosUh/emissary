package middleware

import (
	"io"
	"io/ioutil"
)

type Reverse struct{}

func (reverser *Reverse) Passthru(r io.Reader, w io.Writer) error {
	in, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	length := len(in)

	for i := (length - 1); i >= 0; i-- {

		_, err = w.Write([]byte{in[i]})
		if err != nil {
			return err
		}
	}
	return nil
}
