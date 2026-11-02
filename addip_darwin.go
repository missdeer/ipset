//go:build darwin
// +build darwin

package ipset

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/leiless/dnsredir/pf"
)

var darwinPF = struct {
	sync.Mutex
	dev   int
	users int
}{dev: -1}

// Keep the OS boundary replaceable for tests that must not change the host's PF tables.
var (
	openPF    = pf.Open
	closePF   = pf.Close
	addPFAddr = pf.AddAddr
)

func initLib() error {
	darwinPF.Lock()
	defer darwinPF.Unlock()
	if darwinPF.users == 0 {
		dev, err := openPF(os.O_WRONLY)
		if err != nil {
			return fmt.Errorf("open /dev/pf: %w", err)
		}
		darwinPF.dev = dev
	}
	// CoreDNS can configure ipset in several server blocks sharing this device.
	darwinPF.users++
	return nil
}

func addIP(ip net.IP, list string) error {
	name, anchor := list, ""
	if i := strings.IndexByte(list, ':'); i >= 0 {
		name, anchor = list[:i], list[i+1:]
	}
	tables := make(pf.TableSet)
	if strings.IndexByte(list, 0) >= 0 {
		return fmt.Errorf("invalid PF table name: %q", list)
	}
	if err := tables.Add(name, anchor, 0); err != nil {
		return err
	}
	if ip.To16() == nil {
		return fmt.Errorf("invalid IP address: %v", ip)
	}
	darwinPF.Lock()
	defer darwinPF.Unlock()
	if darwinPF.dev < 0 {
		return fmt.Errorf("PF device is not initialized")
	}
	// Let the kernel deduplicate addresses. A user-space cache would become stale
	// when PF rules/tables are reloaded, preventing a later DNS answer from restoring an IP.
	added, err := addPFAddr(darwinPF.dev, name, anchor, ip)
	if err != nil {
		return fmt.Errorf("add %v to PF table %s: %w", ip, list, err)
	}
	if added {
		AddIPCount.WithLabelValues(list).Inc()
	}
	return nil
}

func flushSet(list string) error {
	// Like the existing BSD implementation, only reset local bookkeeping.
	// Darwin has no address cache; PF table contents belong to the caller.
	return nil
}

func shutdownLib() error {
	darwinPF.Lock()
	defer darwinPF.Unlock()
	if darwinPF.users == 0 {
		return nil
	}
	darwinPF.users--
	if darwinPF.users != 0 {
		return nil
	}
	dev := darwinPF.dev
	darwinPF.dev = -1
	return closePF(dev)
}
