package corednsmikrotik

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

type DHCPLease struct {
	ID               string `json:".id"`
	ActiveAddress    net.IP `json:"active-address"`
	ActiveClientId   string `json:"active-client-id"`
	ActiveMacAddress string `json:"active-mac-address"`
	ActiveServer     string `json:"active-server"`
	Address          net.IP `json:"address"`
	AddressList      string `json:"address-lists"`
	Blocked          string `json:"blocked"`
	ClientId         string `json:"client-id"`
	DHCPOption       string `json:"dhcp-option"`
	Disabled         string `json:"disabled"`
	Dynamic          string `json:"dynamic"`
	ExpiresAfter     string `json:"expires-after"`
	HostName         string `json:"host-name"`
	LastSeen         string `json:"last-seen"`
	MacAddress       string `json:"mac-address"`
	Radius           string `json:"radius"`
	Server           string `json:"server"`
	Status           string `json:"status"`
}

type Client struct {
	log      clog.P
	endpoint *url.URL
	username string
	password string
}

func (c *Client) GetLeases(ctx context.Context) ([]*DHCPLease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint.String(), nil)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(c.username, c.password)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.log.Infof("Could not reach router %s: %s", c.endpoint, err)
		return nil, err
	}
	leases := []*DHCPLease{}
	return leases, json.NewDecoder(res.Body).Decode(&leases)
}
