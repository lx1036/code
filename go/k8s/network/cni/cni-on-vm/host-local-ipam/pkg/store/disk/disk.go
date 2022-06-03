package disk

import (
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strings"

	"path/filepath"
)

const LineBreak = "\n"
const lastIPFilePrefix = "last_reserved_ip."

var defaultDataDir = "/var/lib/cni/networks"

type Store struct {
	*FileLock
	dataDir string
}

func New(name, dataDir string) (*Store, error) {
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	dir := filepath.Join(dataDir, name) // /var/lib/cni/networks/mynet/
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	lk, err := NewFileLock(dir)
	if err != nil {
		return nil, err
	}
	return &Store{lk, dir}, nil
}

func (s *Store) GetByID(id string, ifname string) []net.IP {
	panic("implement me")
}

func (s *Store) Reserve(id string, ifname string, ip net.IP, rangeID string) (bool, error) {
	filename := GetEscapedPath(s.dataDir, ip.String())
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := f.WriteString(strings.TrimSpace(id) + LineBreak + ifname); err != nil {
		f.Close()
		os.Remove(f.Name())
		return false, err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return false, err
	}

	ipfile := GetEscapedPath(s.dataDir, lastIPFilePrefix+rangeID)
	err = ioutil.WriteFile(ipfile, []byte(ip.String()), 0644)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) LastReservedIP(rangeID string) (net.IP, error) {
	panic("implement me")
}

func (s *Store) Release(ip net.IP) error {
	panic("implement me")
}

func (s *Store) ReleaseByID(id string, ifname string) error {
	panic("implement me")
}

func GetEscapedPath(dataDir string, fname string) string {
	if runtime.GOOS == "windows" {
		fname = strings.Replace(fname, ":", "_", -1)
	}
	return filepath.Join(dataDir, fname)
}
