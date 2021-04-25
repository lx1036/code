package cgroups

type Path func(subsystem Name) (string, error)

// StaticPath returns a static path to use for all cgroups
func StaticPath(path string) Path {
	return func(_ Name) (string, error) {
		return path, nil
	}
}
