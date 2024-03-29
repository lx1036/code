package testing

import (
	"net"
	"os"
)

type FakeStore struct {
	ipMap          map[string]string
	lastReservedIP map[string]net.IP
}

func NewFakeStore(ipmap map[string]string, lastIPs map[string]net.IP) *FakeStore {
	return &FakeStore{ipmap, lastIPs}
}

func (s *FakeStore) Lock() error {
	return nil
}

func (s *FakeStore) Unlock() error {
	return nil
}

func (s *FakeStore) Close() error {
	return nil
}

func (s *FakeStore) GetByID(id string, ifname string) []net.IP {
	var ips []net.IP
	for k, v := range s.ipMap {
		if v == id {
			ips = append(ips, net.ParseIP(k))
		}
	}
	return ips
}

func (s *FakeStore) Reserve(id string, ifname string, ip net.IP, rangeID string) (bool, error) {
	key := ip.String()
	if _, ok := s.ipMap[key]; !ok {
		s.ipMap[key] = id
		s.lastReservedIP[rangeID] = ip
		return true, nil
	}
	return false, nil
}

func (s *FakeStore) LastReservedIP(rangeID string) (net.IP, error) {
	ip, ok := s.lastReservedIP[rangeID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return ip, nil
}

func (s *FakeStore) Release(ip net.IP) error {
	panic("implement me")
}

func (s *FakeStore) ReleaseByID(id string, ifname string) error {
	toDelete := []string{}
	for k, v := range s.ipMap {
		if v == id {
			toDelete = append(toDelete, k)
		}
	}
	for _, ip := range toDelete {
		delete(s.ipMap, ip)
	}
	return nil
}
