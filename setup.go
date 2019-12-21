package ipset

import (
	"context"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/caddyserver/caddy"
	"github.com/miekg/dns"
)

// N implements the plugin interface.
type N struct {
	Next     plugin.Handler
	listName []string
}

func init() { plugin.Register("ipset", setup) }

func setup(c *caddy.Controller) error {
	listName := []string{}
	for c.Next() {
		args := c.RemainingArgs()

		copy(listName, args)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return N{Next: next, listName: listName}
	})

	log.Debug("ipset plugin has list:", listName)
	if len(listName) > 0 {
		c.OnStartup(func() error {
			return initLib()
		})

		c.OnShutdown(func() error {
			return shutdownLib()
		})
	}

	return nil
}

// ServeDNS implements the plugin.Handler interface.
func (n N) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	wr := NewResponseReverter(w, r, n.listName)

	if len(n.listName) == 0 {
		log.Debug("no list name")
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}
	log.Debug("has list", n.listName)
	return plugin.NextOrFailure(n.Name(), n.Next, ctx, wr, r)
}

// Name implements the Handler interface.
func (n N) Name() string { return "ipset" }
