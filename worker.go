package corednsmikrotik

import (
	"context"
	"net/url"
	"strings"
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	str2duration "github.com/xhit/go-str2duration/v2"
)

func (c *Mikrotik) start() {
	log := clog.NewWithPlugin("mikrotik[" + c.router + "]")

	log.Infof("starting")
	endpoint := &url.URL{}
	endpoint.Host = c.router
	endpoint.Scheme = "https"
	endpoint.Path = "/rest/ip/dhcp-server/lease"
	c.client = &Client{
		log:      log,
		endpoint: endpoint,
		username: c.username,
		password: c.password,
	}

	c.update()

	go func() {
		ticker := time.NewTicker(c.updateInterval)
		for _ = range ticker.C {
			c.update()
		}
	}()
}

func (c *Mikrotik) update() {
	c.log.Debug("starting update")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	leases, err := c.client.GetLeases(ctx)
	if err != nil {
		c.log.Debugf("Got error: %v", err)
		return
	}

	c.log.Debugf("got %d leases", len(leases))

	update := make(map[string]dns.RR, len(leases))
	for _, lease := range leases {
		lastSeen, err := str2duration.ParseDuration(lease.LastSeen)
		if err != nil {
			c.log.Warningf("Cannot parse last seen: %s", lease.LastSeen)
		} else if lastSeen > c.keep {
			c.log.Debugf("Skipping '%s' (%s), last seen %s", lease.HostName, lease.MacAddress, lastSeen.String())
			continue
		}
		expiresAfter, err := str2duration.ParseDuration(lease.ExpiresAfter)
		if err != nil {
			c.log.Warningf("Cannot parse expires after: %s", lease.ExpiresAfter)
			expiresAfter = 60 * time.Second
		}

		var name string
		if lease.HostName == "" {
			name = strings.ToLower(strings.ReplaceAll(lease.MacAddress, ":", "-")) + "." + c.domain
		} else {
			name = strings.ToLower(lease.HostName) + "." + c.domain
		}
		update[name] = &dns.A{
			Hdr: dns.RR_Header{
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Name:   name,
				Ttl:    uint32(expiresAfter.Seconds()),
			},
			A: lease.Address,
		}
		c.log.Debugf("%s", update[name])
	}

	c.lock.Lock()
	c.cache = update
	c.lock.Unlock()
}
