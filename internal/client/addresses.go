package client

import (
	"fmt"
	"net/url"
)

func (c *Client) GetAddress(id int64) (*Address, error) {
	var a Address
	if err := c.do("GET", fmt.Sprintf("/addresses/%d", id), nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) CreateAddress(in Address) (*Address, error) {
	var a Address
	if err := c.do("POST", "/addresses", in, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) UpdateAddress(id int64, in Address) (*Address, error) {
	var a Address
	if err := c.do("PUT", fmt.Sprintf("/addresses/%d", id), in, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) DeleteAddress(id int64, cleanupKeys []string) error {
	body := map[string]any{}
	if len(cleanupKeys) > 0 {
		body["cleanup_keys"] = cleanupKeys
	}
	return c.do("DELETE", fmt.Sprintf("/addresses/%d", id), body, nil)
}

func (c *Client) GetAddressByIP(ip string) (*Address, error) {
	var a Address
	if err := c.do("GET", "/addresses/by-ip/"+url.PathEscape(ip), nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// DeletePreviewKeys returns the provider-cleanup keys for an address delete.
func (c *Client) DeletePreviewKeys(id int64) ([]string, error) {
	var p deletePreview
	if err := c.do("GET", fmt.Sprintf("/addresses/%d/delete-preview", id), nil, &p); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(p.Items))
	for _, it := range p.Items {
		keys = append(keys, it.Key)
	}
	return keys, nil
}
