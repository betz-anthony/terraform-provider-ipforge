package client

import (
	"fmt"
	"net/url"
)

func (c *Client) CreateDNSRecord(zone string, in DNSRecord) (*DNSRecord, error) {
	var r DNSRecord
	if err := c.do("POST", fmt.Sprintf("/dns/zones/%s/records", url.PathEscape(zone)), in, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) DeleteDNSRecord(zone string, rec DNSRecord) error {
	return c.do("DELETE", fmt.Sprintf("/dns/zones/%s/records", url.PathEscape(zone)), rec, nil)
}

// FindDNSRecord reads a zone's records (paginated) and returns the one matching
// name+type+value, or nil if absent.
func (c *Client) FindDNSRecord(zone, name, rtype, value string) (*DNSRecord, error) {
	offset := int64(0)
	for {
		var env pageEnvelope[DNSRecord]
		path := fmt.Sprintf("/dns/zones/%s/records?limit=200&offset=%d", url.PathEscape(zone), offset)
		if err := c.do("GET", path, nil, &env); err != nil {
			return nil, err
		}
		for i := range env.Items {
			r := env.Items[i]
			if r.Name == name && r.RecordType == rtype && r.Value == value {
				return &r, nil
			}
		}
		offset += int64(len(env.Items))
		if len(env.Items) == 0 || offset >= env.Total {
			return nil, nil
		}
	}
}
