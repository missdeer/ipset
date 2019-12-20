// +build linux

package ipset

import (
	goipset "github.com/digineo/go-ipset/v2"
	"github.com/mdlayher/netlink"
	"github.com/ti-mo/netfilter"
	"net"
)

var (
	c *goipset.Conn
)

func initLib() (err error) {
	c, err = goipset.Dial(netfilter.ProtoIPv4, &netlink.Config{})
	return
}

func addIP(ip net.IP, list string) error {
	return c.Add(list, goipset.NewEntry(goipset.EntryIP(ip)))
}
