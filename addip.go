// +build !linux

package ipset

import "net"

func initLib() {

}

func addIP(ip net.IP, list string) error {
	return nil
}
