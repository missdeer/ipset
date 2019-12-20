// +build linux

package ipset

import (
	"github.com/ti-mo/netfilter"
	ipset "gopkg.in/digineo/go-ipset.v2"
	"net"
)

var (
	c ipset.Conn
)

func initLib() {
	c = ipset.Conn{ipset.Family: netfilter.ProtoIPv4}
}

func addIP(ip net.IP, list string) error {
	c.Add(list, NewEntry(EntryIP(ip)))
	return nil
}
