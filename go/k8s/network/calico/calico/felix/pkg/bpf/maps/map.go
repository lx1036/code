package maps

import "fmt"

type MapParameters struct {
	Filename     string
	Type         string
	KeySize      int
	ValueSize    int
	MaxEntries   int
	Name         string
	Flags        int
	Version      int
	UpdatedByBPF bool
}

func versionedStr(ver int, str string) string {
	if ver <= 1 {
		return str
	}

	return fmt.Sprintf("%s%d", str, ver)
}

func (mp *MapParameters) VersionedName() string {
	return versionedStr(mp.Version, mp.Name)
}

func (mp *MapParameters) VersionedFilename() string {
	return versionedStr(mp.Version, mp.Filename)
}
