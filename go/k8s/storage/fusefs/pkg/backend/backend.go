package backend

import "io"

type Capabilities struct {
	NoParallelMultipart bool
	MaxMultipartSize    uint64
	// indicates that the blob store has native support for directories
	DirBlob bool
	Name    string
}

type Backend interface {
	Write(file string, offset int64, data []byte) (wsize int, err error)
	Read(file string, offset int64, data []byte) (rsize int, err error)
	WriteStream(file string, offset int64, length int64, reader io.ReadSeeker) (wsize int,
		err error)
	ReadStream(file string, offset int64, length int64, writer io.Writer) (rsize int,
		err error)
	WriteStreamWithCallBack(file string, offset int64, length int64, reader io.ReadSeeker,
		cb IOCallback)
	ReadStreamWithCallBack(file string, offset int64, length int64, writer io.Writer,
		cb IOCallback)
	WriteV(file string, vec *IOVector, reader io.ReadSeeker) (wsize int, err error)
	ReadV(file string, vec *IOVector, writer io.Writer) (rsize int, err error)
	Truncate(file string, offset int64) (err error)
	Fallocate(file string, op int, off int64, len int64) (err error)
	Flush(file string) (err error)
	Rename(src string, dst string, dir bool) (err error)
	Delete(file string) (err error)
	Deletes(files []string) (err error)
	SupportCallBack() bool
}

type IOVector struct {
	Ranges []*Range
}

type Range struct {
	Offset int64
	Length int64
}

type IOCallback interface {
	SetError(error)
	Run()
}
