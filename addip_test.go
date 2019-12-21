package ipset

import (
	"net"
	"testing"
)

func TestAddIP(t *testing.T) {
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
