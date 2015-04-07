// Middleware for encrypting the data with the provided
// GPG public key.

package pgp

import (
	"github.com/maxwellhealth/go-gpg"
	"io"
)

type PGP struct {
	Key []byte
}

func (p *PGP) Passthru(r io.Reader, w io.Writer) error {
	return gpg.Encode(p.Key, r, w)
}
