package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// PossibleCPUSysfsPath is used to retrieve the number of CPUs for per-CPU maps.
	PossibleCPUSysfsPath = "/sys/devices/system/cpu/possible"
)

// GetNumPossibleCPUs returns a total number of possible CPUS, i.e. CPUs that
// have been allocated resources and can be brought online if they are present.
// The number is retrieved by parsing /sys/device/system/cpu/possible.
//
// See https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/linux/cpumask.h?h=v4.19#n50
// for more details.
func GetNumPossibleCPUs() int {
	f, err := os.Open(PossibleCPUSysfsPath)
	if err != nil {
		log.WithError(err).Errorf("unable to open %q", PossibleCPUSysfsPath)
		return 0
	}
	defer f.Close()

	return getNumPossibleCPUsFromReader(f)
}

func getNumPossibleCPUsFromReader(r io.Reader) int {
	out, err := ioutil.ReadAll(r)
	if err != nil {
		log.WithError(err).Errorf("unable to read %q to get CPU count", PossibleCPUSysfsPath)
		return 0
	}

	var start, end int
	count := 0
	for _, s := range strings.Split(string(out), ",") {
		// Go's scanf will return an error if a format cannot be fully matched.
		// So, just ignore it, as a partial match (e.g. when there is only one
		// CPU) is expected.
		n, err := fmt.Sscanf(s, "%d-%d", &start, &end) // 0-23

		switch n {
		case 0:
			log.WithError(err).Errorf("failed to scan %q to retrieve number of possible CPUs!", s)
			return 0
		case 1:
			count++
		default:
			count += (end - start + 1)
		}
	}

	return count // 24 个核
}
