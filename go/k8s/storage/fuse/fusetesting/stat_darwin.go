package fusetesting

import (
	"syscall"
	"time"
)

func extractMtime(sys interface{}) (mtime time.Time, ok bool) {
	return time.Unix(sys.(*syscall.Stat_t).Mtimespec.Unix()), true
}

func extractBirthtime(sys interface{}) (birthtime time.Time, ok bool) {
	return time.Unix(sys.(*syscall.Stat_t).Birthtimespec.Unix()), true
}

func extractNlink(sys interface{}) (nlink uint64, ok bool) {
	return uint64(sys.(*syscall.Stat_t).Nlink), true
}

func getTimes(stat *syscall.Stat_t) (atime, ctime, mtime time.Time) {
	atime = time.Unix(stat.Atimespec.Unix())
	ctime = time.Unix(stat.Ctimespec.Unix())
	mtime = time.Unix(stat.Mtimespec.Unix())
	return atime, ctime, mtime
}
