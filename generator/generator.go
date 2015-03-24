package generator

import (
	"io"
)

type FileGenerator interface {
	Generate(io.Writer) error
}
