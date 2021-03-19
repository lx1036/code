package bolt

type pgid uint64

// page 是
type page struct {
	id       pgid
	flags    uint16
	count    uint16
	overflow uint32
}

type pages []*page

func (s pages) Len() int           { return len(s) }
func (s pages) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pages) Less(i, j int) bool { return s[i].id < s[j].id }
