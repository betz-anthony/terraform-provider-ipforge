package client

import "fmt"

func (c *Client) GetVlan(id int64) (*Vlan, error) {
	var v Vlan
	if err := c.do("GET", fmt.Sprintf("/vlans/%d", id), nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
func (c *Client) CreateVlan(in Vlan) (*Vlan, error) {
	var v Vlan
	if err := c.do("POST", "/vlans", in, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
func (c *Client) UpdateVlan(id int64, in Vlan) (*Vlan, error) {
	var v Vlan
	if err := c.do("PUT", fmt.Sprintf("/vlans/%d", id), in, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
func (c *Client) DeleteVlan(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/vlans/%d", id), nil, nil)
}
