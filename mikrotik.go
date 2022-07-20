package corednsmikrotik

import (
	"context"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("mikrotik")

type Mikrotik struct {
	stop   context.Context
	cancel context.CancelFunc

	Next plugin.Handler
	config

	log    clog.P
	lock   sync.RWMutex
	client *Client
}

type config struct {
	router         string
	username       string
	password       string
	domain         string
	cache          map[string]dns.RR
	updateInterval time.Duration
	keep           time.Duration
}

func New(next plugin.Handler, config config) *Mikrotik {
	stop, cancel := context.WithCancel(context.Background())
	c := &Mikrotik{
		stop:   stop,
		cancel: cancel,
		Next:   next,
		log:    clog.NewWithPlugin("mikrotik[" + config.router + "]"),
		config: config,
	}
	c.start()
	return c
}

func (c *Mikrotik) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{Req: r, W: w}
	m := new(dns.Msg)
	m.SetReply(r)

	name := state.Name()
	c.lock.RLock()
	rr, ok := c.cache[name]
	c.lock.RUnlock()

	if ok {
		m.Authoritative, m.RecursionAvailable, m.Compress = true, true, true
		m.Answer = []dns.RR{rr}
	}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

func (*Mikrotik) Name() string {
	return "mikrotik"
}
