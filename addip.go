// +build !linux

package ipset

import "net"

func initLib() error {
	return nil
}

func addIP(ip net.IP, list string) error {
	return nil
}
