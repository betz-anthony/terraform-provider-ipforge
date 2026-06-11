package client

import "fmt"

func (c *Client) Allocate(subnetID int64, in AllocateRequest) (*AllocateResult, error) {
	var r AllocateResult
	if err := c.do("POST", fmt.Sprintf("/subnets/%d/allocate", subnetID), in, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
