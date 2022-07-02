package datapath

import (
	"io"
)

type ConfigWriter interface {
	WriteNodeConfig(io.Writer, *LocalNodeConfiguration) error
}
