package ipset

import (
	"net"
	"os"
	"testing"
)

func TestAddIP(t *testing.T) {
	if os.Getenv("IPSET_INTEGRATION_TEST") != "1" {
		t.Skip("set IPSET_INTEGRATION_TEST=1 to modify the host's testset table")
	}
	err := initLib()
	if err != nil {
		t.Error(err)
	}
	err = flushSet("testset")
	if err != nil {
		t.Error(err)
	}
	err = addIP(net.ParseIP("10.100.100.10"), "testset")
	if err != nil {
		t.Error(err)
	}
	err = shutdownLib()
	if err != nil {
		t.Error(err)
	}
}
