package client

import (
	"fmt"
	"net/url"
)

func (c *Client) AddReservation(scopeID string, in DHCPLease) (*DHCPLease, error) {
	var l DHCPLease
	if err := c.do("POST", fmt.Sprintf("/dhcp/scopes/%s/reservations", url.PathEscape(scopeID)), in, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (c *Client) DeleteReservation(scopeID, ip string) error {
	return c.do("DELETE", fmt.Sprintf("/dhcp/scopes/%s/reservations/%s", url.PathEscape(scopeID), url.PathEscape(ip)), nil, nil)
}

func (c *Client) FindReservation(scopeID, ip string) (*DHCPLease, error) {
	offset := int64(0)
	for {
		var env pageEnvelope[DHCPLease]
		path := fmt.Sprintf("/dhcp/scopes/%s/leases?limit=200&offset=%d", url.PathEscape(scopeID), offset)
		if err := c.do("GET", path, nil, &env); err != nil {
			return nil, err
		}
		for i := range env.Items {
			if env.Items[i].IPAddress == ip {
				return &env.Items[i], nil
			}
		}
		offset += int64(len(env.Items))
		if len(env.Items) == 0 || offset >= env.Total {
			return nil, nil
		}
	}
}
