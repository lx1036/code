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

func (s *Store) Reserve(id string, ifname string, ip net.IP, rangeID string) (bool, error) {
	f, err := os.OpenFile(s.getIPFile(ip), os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := f.WriteString(s.getIPFileContent(id, ifname)); err != nil {
		f.Close()
		os.Remove(f.Name())
		return false, err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return false, err
	}

	ipfile := s.getLastReservedIPFile(rangeID)
	err = ioutil.WriteFile(ipfile, []byte(ip.String()), 0644)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) GetByID(id string, ifname string) []net.IP {
	var ips []net.IP

	match := s.getIPFileContent(id, ifname)
	// walk through all ips in this network to get the ones which belong to a specific ID
	_ = filepath.Walk(s.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.TrimSpace(string(data)) == match {
			_, ipString := filepath.Split(path) // INFO: ip 是文件名，做的不太好
			if ip := net.ParseIP(ipString); ip != nil {
				ips = append(ips, ip)
			}
		}

		return nil
	})

	return ips
}

func (s *Store) ReleaseByID(id string, ifname string) error {
	match := s.getIPFileContent(id, ifname)
	err := filepath.Walk(s.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.TrimSpace(string(data)) == match {
			if err := os.Remove(path); err != nil {
				return nil
			}
		}

		return nil
	})

	return err
}

func (s *Store) getIPFileContent(id string, ifname string) string {
	return strings.TrimSpace(id) + LineBreak + ifname
}

func (s *Store) LastReservedIP(rangeID string) (net.IP, error) {
	ipfile := s.getLastReservedIPFile(rangeID)
	data, err := ioutil.ReadFile(ipfile)
	if err != nil {
		return nil, err
	}

	return net.ParseIP(string(data)), nil
}

func (s *Store) Release(ip net.IP) error {
	return os.Remove(s.getIPFile(ip))
}

func (s *Store) getLastReservedIPFile(rangeID string) string {
	return GetEscapedPath(s.dataDir, lastIPFilePrefix+rangeID)
}

func (s *Store) getIPFile(ip net.IP) string {
	return GetEscapedPath(s.dataDir, ip.String())
}

func GetEscapedPath(dataDir string, fname string) string {
	if runtime.GOOS == "windows" {
		fname = strings.Replace(fname, ":", "_", -1)
	}
	return filepath.Join(dataDir, fname)
}
