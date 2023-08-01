package ipset

import (
	"context"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/coredns/caddy"
	"github.com/miekg/dns"
)

// N implements the plugin interface.
type N struct {
	Next           plugin.Handler
	listName       []string
	mappedListName map[string][]string
}

func init() { plugin.Register("ipset", setup) }

func setup(c *caddy.Controller) error {
	mappedListName := make(map[string][]string)
	listName := []string{}
	for c.Next() {
		args := c.RemainingArgs()
		hasBlock := false
		for c.NextBlock() {
			val := c.Val()
			blockArgs := []string{}
			for _, arg := range c.RemainingArgs() {
				blockArgs = append(blockArgs, plugin.Host(arg).NormalizeExact()...)
			}
			if val != "" && len(blockArgs) > 0 {
				mappedListName[val] = blockArgs
				hasBlock = true
			}
		}
		if !hasBlock {
			listName = append(listName, args...)
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return N{Next: next, listName: listName, mappedListName: mappedListName}
	})

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
	wr := NewResponseReverter(w, r, n.listName, n.mappedListName)

	if len(n.listName) == 0 {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}
	return plugin.NextOrFailure(n.Name(), n.Next, ctx, wr, r)
}

// Name implements the Handler interface.
func (n N) Name() string { return "ipset" }
