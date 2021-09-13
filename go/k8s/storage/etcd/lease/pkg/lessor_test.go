package pkg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	minLeaseTTL         = int64(5)
	minLeaseTTLDuration = time.Duration(minLeaseTTL) * time.Second
)

func NewTestBackend(t *testing.T) (string, backend.Backend) {
	tmpPath, err := ioutil.TempDir("lease-test", "lease")
	if err != nil {
		t.Fatalf("failed to create tmpdir (%v)", err)
	}
	bcfg := backend.DefaultBackendConfig()
	bcfg.Path = filepath.Join(tmpPath, "be")
	return tmpPath, backend.New(bcfg)
}

// TestLeaseConcurrentKeys ensures Lease.Keys method calls are guarded
// from concurrent map writes on 'itemSet'.
func TestLeaseConcurrentKeys(t *testing.T) {
	dir, be := NewTestBackend(t)
	defer os.RemoveAll(dir)
	defer be.Close()

	le := NewLessor(be, LessorConfig{MinLeaseTTL: minLeaseTTL})
	defer le.Stop()

	// grant a lease with long term (100 seconds) to
	// avoid early termination during the test.
	lease, err := le.Grant(1, 100)
	if err != nil {
		t.Fatalf("could not grant lease for 100s ttl (%v)", err)
	}

}
