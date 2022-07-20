package corednsmikrotik

import (
	"strings"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

func init() { plugin.Register("mikrotik", setup) }

func getDuration(d *time.Duration, c *caddy.Controller) bool {
	var tmp string
	if !c.Args(&tmp) {
		return false
	}
	updateInterval, err := time.ParseDuration(tmp)
	if err != nil {
		return false
	}
	*d = updateInterval
	return true
}

func setup(c *caddy.Controller) error {
	var log = clog.NewWithPlugin("mikrotik.setup")
	c.Next() // skip name

	config := config{
		updateInterval: 5 * time.Second,
		keep:           5 * time.Minute,
	}

	routers := c.RemainingArgs()
	if len(routers) != 1 {
		return c.ArgErr()
	}
	config.router = routers[0]

	if c.NextBlock() {
		for {
			key := c.Val()
			switch key {
			case "domain":
				if !c.Args(&config.domain) {
					return c.Errf("domain needs one argument exactly")
				}
				if !strings.HasSuffix(config.domain, ".") {
					config.domain += "."
				}
			case "username":
				if !c.Args(&config.username) {
					return c.Errf("username needs one argument exactly")
				}
			case "password":
				if !c.Args(&config.password) {
					return c.Errf("password needs one argument exactly")
				}
			case "keep":
				if !getDuration(&config.keep, c) {
					return c.Errf("keep needs one time.Duration argument exactly")
				}
			case "update":
				if !getDuration(&config.updateInterval, c) {
					return c.Errf("upate needs one time.Duration argument exactly")
				}
			case "fallthrough":
				config.fall.SetZonesFromArgs(c.RemainingArgs())
			}
			if !c.NextLine() {
				break
			}
		}
	}
	if config.domain == "" {
		log.Info("no domain")
	}
	if config.username == "" {
		return c.Errf("no username")
	}
	if config.password == "" {
		return c.Errf("no password")
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return New(next, config)
	})

	return nil
}
