package client

import "fmt"

func (c *Client) GetSubnet(id int64) (*Subnet, error) {
	var s Subnet
	if err := c.do("GET", fmt.Sprintf("/subnets/%d", id), nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) CreateSubnet(in Subnet) (*Subnet, error) {
	var s Subnet
	if err := c.do("POST", "/subnets", in, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) UpdateSubnet(id int64, in Subnet) (*Subnet, error) {
	var s Subnet
	if err := c.do("PUT", fmt.Sprintf("/subnets/%d", id), in, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) DeleteSubnet(id int64) error {
	return c.do("DELETE", fmt.Sprintf("/subnets/%d", id), nil, nil)
}

func (c *Client) ListSubnets() ([]Subnet, error) {
	var out []Subnet
	if err := c.do("GET", "/subnets", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
