//go:build darwin
// +build darwin

package ipset

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func mockDarwinPF(t *testing.T) func() {
	t.Helper()
	oldOpen, oldClose, oldAdd := openPF, closePF, addPFAddr
	darwinPF.dev, darwinPF.users = -1, 0
	openPF = func(flags int) (int, error) {
		if flags != os.O_WRONLY {
			t.Errorf("open flags = %d", flags)
		}
		return 42, nil
	}
	closePF = func(dev int) error {
		if dev != 42 {
			t.Errorf("close device = %d", dev)
		}
		return nil
	}
	addPFAddr = func(dev int, name, anchor string, ip net.IP) (bool, error) { return true, nil }
	return func() {
		openPF, closePF, addPFAddr = oldOpen, oldClose, oldAdd
		darwinPF.dev, darwinPF.users = -1, 0
	}
}

func TestDarwinPFSharedDevice(t *testing.T) {
	defer mockDarwinPF(t)()
	opens, closes := 0, 0
	openPF = func(int) (int, error) { opens++; return 42, nil }
	closePF = func(int) error { closes++; return nil }
	for i := 0; i < 2; i++ {
		if err := initLib(); err != nil {
			t.Fatal(err)
		}
	}
	if opens != 1 {
		t.Fatalf("opened %d devices for two server blocks", opens)
	}
	if err := shutdownLib(); err != nil || closes != 0 {
		t.Fatalf("closed while another server block is active: %v", err)
	}
	if err := addIP(net.ParseIP("192.0.2.1"), "direct_cidr"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := shutdownLib(); err != nil {
			t.Fatal(err)
		}
	}
	if closes != 1 || darwinPF.dev != -1 {
		t.Fatalf("shutdown: closes=%d device=%d", closes, darwinPF.dev)
	}
	if err := addIP(net.ParseIP("192.0.2.1"), "direct_cidr"); err == nil {
		t.Fatal("write after shutdown succeeded")
	}
}

func TestDarwinPFAnchorsAndRetry(t *testing.T) {
	defer mockDarwinPF(t)()
	if err := initLib(); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("ioctl failed")
	calls := 0
	addPFAddr = func(dev int, name, anchor string, ip net.IP) (bool, error) {
		calls++
		if dev != 42 || name != "direct_cidr" || anchor != "com.apple/climbovergui" || !ip.Equal(net.ParseIP("192.0.2.1")) {
			t.Fatalf("unexpected PF write: %d %q %q %v", dev, name, anchor, ip)
		}
		if calls == 1 {
			return false, failure
		}
		return false, nil // Already present is a successful operation.
	}
	for i := 0; i < 3; i++ {
		err := addIP(net.ParseIP("192.0.2.1"), "direct_cidr:com.apple/climbovergui")
		if (i == 0 && !errors.Is(err, failure)) || (i > 0 && err != nil) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	if calls != 3 {
		t.Fatal("cached IP prevented retry or restoration after a PF reload")
	}
	addPFAddr = func(dev int, name, anchor string, ip net.IP) (bool, error) {
		if name != "appleroute" || anchor != "" || ip.To4() != nil {
			t.Fatalf("unexpected IPv6/global table write: %q %q %v", name, anchor, ip)
		}
		return true, nil
	}
	if err := addIP(net.ParseIP("2001:db8::1"), "appleroute"); err != nil {
		t.Fatal(err)
	}
}

func TestDarwinPFValidationAndOpenFailure(t *testing.T) {
	defer mockDarwinPF(t)()
	failure := errors.New("permission denied")
	openPF = func(int) (int, error) { return 0, failure }
	if err := initLib(); !errors.Is(err, failure) || darwinPF.users != 0 || darwinPF.dev != -1 {
		t.Fatalf("failed startup retained device state: %v", err)
	}
	openPF = func(int) (int, error) { return 42, nil }
	if err := initLib(); err != nil {
		t.Fatal(err)
	}
	addPFAddr = func(int, string, string, net.IP) (bool, error) {
		t.Fatal("invalid input reached PF")
		return false, nil
	}
	for _, list := range []string{"", ":anchor", "table\x00suffix", strings.Repeat("a", 32), "table:" + strings.Repeat("a", 1024)} {
		if err := addIP(net.ParseIP("192.0.2.1"), list); err == nil {
			t.Errorf("accepted invalid table %q", list)
		}
	}
	if err := addIP(nil, "table"); err == nil {
		t.Fatal("accepted invalid IP")
	}
}

type darwinDNSReply struct{}

func (darwinDNSReply) Name() string { return "darwin-test" }

func (darwinDNSReply) ServeDNS(_ context.Context, w dns.ResponseWriter, request *dns.Msg) (int, error) {
	reply := new(dns.Msg)
	reply.SetReply(request)
	reply.Answer = []dns.RR{&dns.A{
		Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
		A:   net.ParseIP("192.0.2.9"),
	}}
	return dns.RcodeSuccess, w.WriteMsg(reply)
}

func TestDarwinDNSDomainMapping(t *testing.T) {
	defer mockDarwinPF(t)()
	if err := initLib(); err != nil {
		t.Fatal(err)
	}
	calls := 0
	addPFAddr = func(dev int, name, anchor string, ip net.IP) (bool, error) {
		calls++
		if name != "direct_cidr" || anchor != "com.apple/climbovergui" || !ip.Equal(net.ParseIP("192.0.2.9")) {
			t.Fatalf("unexpected DNS-to-PF write: %q %q %v", name, anchor, ip)
		}
		return true, nil
	}
	handler := N{Next: darwinDNSReply{}, mappedListName: map[string][]string{
		"direct_cidr:com.apple/climbovergui": {"example.test."},
	}}
	for _, domain := range []string{"sub.example.test.", "unrelated.test."} {
		request := new(dns.Msg)
		request.SetQuestion(domain, dns.TypeA)
		writer := new(test.ResponseWriter)
		if _, err := handler.ServeDNS(context.Background(), writer, request); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected only the matching subdomain to write PF, got %d writes", calls)
	}
}

func TestDarwinPFConcurrentWrites(t *testing.T) {
	defer mockDarwinPF(t)()
	if err := initLib(); err != nil {
		t.Fatal(err)
	}
	var calls int
	addPFAddr = func(int, string, string, net.IP) (bool, error) { calls++; return false, nil }
	var group sync.WaitGroup
	for i := 0; i < 16; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := addIP(net.ParseIP("192.0.2.1"), "table"); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if calls != 16 {
		t.Fatalf("lost concurrent writes: %d", calls)
	}
}
