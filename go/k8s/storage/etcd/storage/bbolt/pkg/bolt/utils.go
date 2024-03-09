package bolt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// The time elapsed between consecutive file locking attempts.
const flockRetryTimeout = 50 * time.Millisecond

var (
	ErrTimeout         = errors.New("timeout")
	ErrInvalid         = errors.New("invalid database")
	ErrVersionMismatch = errors.New("version mismatch")
	ErrChecksum        = errors.New("checksum error")
)

// flock acquires an advisory lock on a file descriptor.
func flock(db *DB, exclusive bool, timeout time.Duration) error {
	var t time.Time
	if timeout != 0 {
		t = time.Now()
	}

	fd := db.file.Fd()

	flag := syscall.LOCK_NB
	if exclusive {
		flag |= syscall.LOCK_EX
	} else {
		flag |= syscall.LOCK_SH
	}

	for {
		// Attempt to obtain an exclusive lock.
		err := syscall.Flock(int(fd), flag)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK {
			return err
		}

		// If we timed out then return an error.
		if timeout != 0 && time.Since(t) > timeout-flockRetryTimeout {
			return ErrTimeout
		}

		// Wait for a bit and try again.
		time.Sleep(flockRetryTimeout)
	}
}

// munmap unmaps a DB's data file from memory.
func munmap(db *DB) error {
	// Ignore the unmap if we have no mapped data.
	if db.dataref == nil {
		return nil
	}

	// Unmap using the original byte slice.
	err := syscall.Munmap(db.dataref)
	db.dataref = nil
	db.data = nil
	db.datasz = 0
	return err
}

// funlock releases an advisory lock on a file descriptor.
func funlock(db *DB) error {
	return syscall.Flock(int(db.file.Fd()), syscall.LOCK_UN)
}

// maxAllocSize is the size used when creating array pointers.
const maxAllocSize = 0x7FFFFFFF

func unsafeAdd(base unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(base) + offset)
}

// INFO: 申请字节空间，取字节数据，base+offset开始，取这段[i,j]字节数据
func unsafeByteSlice(base unsafe.Pointer, offset uintptr, i, j int) []byte {
	return (*[maxAllocSize]byte)(unsafeAdd(base, offset))[i:j:j]
}

// _assert will panic with a given formatted message if the given condition is false.
func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

// ////////// DBWrapper for test ////////
type DBWrapper struct {
	*DB
}

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

// MustOpenDB returns a new, open DB at a temporary location.
func MustOpenDB() *DBWrapper {
	db, err := Open(tempfile(), 0666, nil)
	if err != nil {
		panic(err)
	}
	return &DBWrapper{DB: db}
}

// MustClose closes the database and deletes the underlying file. Panic on error.
func (db *DBWrapper) MustClose() {
	if err := db.Close(); err != nil {
		panic(err)
	}
}

// Close closes the database and deletes the underlying file.
func (db *DBWrapper) Close() error {
	// Check database consistency after every test.
	db.MustCheck()

	// Close database and remove file.
	defer os.Remove(db.Path())
	return db.Close()
}

// MustCheck runs a consistency check on the database and panics if any errors are found.
func (db *DBWrapper) MustCheck() {
	if err := db.Update(func(tx *Tx) error {
		// Collect all the errors.
		var errors []error
		for err := range tx.Check() {
			errors = append(errors, err)
			if len(errors) > 10 {
				break
			}
		}

		// If errors occurred, copy the DB and print the errors.
		// 如果数据一致性错误
		if len(errors) > 0 {
			var path = tempfile()
			if err := tx.CopyFile(path, 0600); err != nil {
				panic(err)
			}

			// Print errors.
			fmt.Print("\n\n")
			fmt.Printf("consistency check failed (%d errors)\n", len(errors))
			for _, err := range errors {
				fmt.Println(err)
			}
			fmt.Println("")
			fmt.Println("db saved to:")
			fmt.Println(path)
			fmt.Print("\n\n")
			os.Exit(-1)
		}

		return nil
	}); err != nil && err != ErrDatabaseNotOpen {
		panic(err)
	}
}
