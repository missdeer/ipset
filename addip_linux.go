// +build linux

package ipset

import (
	"errors"
	"net"

	goipset "github.com/digineo/go-ipset/v2"
	"github.com/mdlayher/netlink"
	"github.com/ti-mo/netfilter"
)

var (
	c *goipset.Conn
)

func initLib() (err error) {
	c, err = goipset.Dial(netfilter.ProtoUnspec, &netlink.Config{})
	return
}

func addIP(ip net.IP, list string) error {
	p, err := c.Header(list)
	if err != nil {
		return err
	}
	var typeMatch bool
	if uint(p.Family.Value) == uint(netfilter.ProtoIPv4) {
		typeMatch = ip.To4() != nil
	} else if uint(p.Family.Value) == uint(netfilter.ProtoIPv6) {
		typeMatch = ip.To16() != nil
	}
	if !typeMatch {
		return errors.New("not matched type")
	}
	AddIPCount.WithLabelValues(list).Add(1)
	return c.Add(list, goipset.NewEntry(goipset.EntryIP(ip)))
}

func flushSet(list string) error {
	return c.Flush(list)
}

func shutdownLib() error {
	return c.Close()
}
