package master

import (
	"fmt"
)

func keyNotFound(name string) (err error) {
	return fmt.Errorf("parameter %v not found", name)
}
