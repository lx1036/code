package store

import "net"

type Store interface {
	Lock() error
	Unlock() error
	Close() error

	GetByID(id string, ifname string) []net.IP
	Reserve(id string, ifname string, ip net.IP, rangeID string) (bool, error)
	LastReservedIP(rangeID string) (net.IP, error)
	Release(ip net.IP) error
	ReleaseByID(id string, ifname string) error
}
