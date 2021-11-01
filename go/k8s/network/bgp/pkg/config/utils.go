package config

import "fmt"

func mapkey(index int, name string) string {
	if name != "" {
		return name
	}

	return fmt.Sprintf("%v", index)
}
