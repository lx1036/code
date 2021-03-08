package backend

type Capabilities struct {
	NoParallelMultipart bool
	MaxMultipartSize    uint64
	// indicates that the blob store has native support for directories
	DirBlob bool
	Name    string
}
