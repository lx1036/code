package fusetesting

import (
	"syscall"
	"time"
)

func extractMtime(sys interface{}) (mtime time.Time, ok bool) {
	return time.Unix(sys.(*syscall.Stat_t).Mtim.Unix()), true
}

func extractBirthtime(sys interface{}) (birthtime time.Time, ok bool) {
	return time.Time{}, false
}

func extractNlink(sys interface{}) (nlink uint64, ok bool) {
	return sys.(*syscall.Stat_t).Nlink, true
}

func getTimes(stat *syscall.Stat_t) (atime, ctime, mtime time.Time) {
	atime = time.Unix(stat.Atim.Unix())
	ctime = time.Unix(stat.Ctim.Unix())
	mtime = time.Unix(stat.Mtim.Unix())
	return atime, ctime, mtime
}
