package multiraft

func Min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
