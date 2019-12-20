// +build !linux

package ipset

import "net"

func initLib() error {}

func addIP(ip net.IP, list string) error {
	return nil
}
