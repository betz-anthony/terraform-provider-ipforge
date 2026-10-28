package client

import (
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) ListPools(subnetID int64) ([]SubnetPool, error) {
	var out []SubnetPool
	if err := c.do("GET", fmt.Sprintf("/subnets/%d/pools", subnetID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetPool finds a pool by id. The API has no single-pool GET, so this reads
// the subnet's list; a missing pool is reported as a 404 APIError so callers
// can use IsNotFound.
func (c *Client) GetPool(subnetID, poolID int64) (*SubnetPool, error) {
	pools, err := c.ListPools(subnetID)
	if err != nil {
		return nil, err
	}
	for i := range pools {
		if pools[i].ID == poolID {
			return &pools[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Detail: "pool not found"}
}

// FindPoolByName finds a pool by name (case-insensitive: the server stores
// names lowercased). A missing pool is a 404 APIError.
func (c *Client) FindPoolByName(subnetID int64, name string) (*SubnetPool, error) {
	pools, err := c.ListPools(subnetID)
	if err != nil {
		return nil, err
	}
	for i := range pools {
		if strings.EqualFold(pools[i].Name, name) {
			return &pools[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Detail: "pool not found"}
}

func (c *Client) CreatePool(subnetID int64, in SubnetPoolInput) (*SubnetPool, error) {
	var p SubnetPool
	if err := c.do("POST", fmt.Sprintf("/subnets/%d/pools", subnetID), in, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) UpdatePool(subnetID, poolID int64, in SubnetPoolInput) (*SubnetPool, error) {
	var p SubnetPool
	if err := c.do("PUT", fmt.Sprintf("/subnets/%d/pools/%d", subnetID, poolID), in, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) DeletePool(subnetID, poolID int64) error {
	return c.do("DELETE", fmt.Sprintf("/subnets/%d/pools/%d", subnetID, poolID), nil, nil)
}
