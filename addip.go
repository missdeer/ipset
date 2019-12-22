// +build !linux

package ipset

import "net"

func initLib() error {
	log.Debug("init ipset lib")
	return nil
}

func addIP(ip net.IP, list string) error {
	log.Debug("add IP:", ip, " to ipset:", list)
	return nil
}

func shutdownLib() error {
	log.Debug("shutdown ipset lib")
	return nil
}