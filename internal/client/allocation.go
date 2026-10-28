package client

import (
	"fmt"
	"strings"
)

// PoolNotSupportedError means a pool was requested but the server answered
// without it, i.e. it predates allocation pools and took the address from the
// whole subnet. The address is already allocated server-side.
type PoolNotSupportedError struct {
	Pool    string
	Address string
}

func (e *PoolNotSupportedError) Error() string {
	return fmt.Sprintf(
		"server did not allocate from pool %q (it may not support allocation pools; upgrade IPForge); it allocated %s from the whole subnet instead",
		e.Pool, e.Address)
}

func (c *Client) Allocate(subnetID int64, in AllocateRequest) (*AllocateResult, error) {
	var r AllocateResult
	if err := c.do("POST", fmt.Sprintf("/subnets/%d/allocate", subnetID), in, &r); err != nil {
		return nil, err
	}
	// Request bodies tolerate unknown fields, so an older server silently
	// ignores `pool`. Never report success for an address that did not come
	// from the requested pool.
	if in.Pool != "" && !strings.EqualFold(r.Pool, in.Pool) {
		return nil, &PoolNotSupportedError{Pool: in.Pool, Address: r.Address}
	}
	return &r, nil
}
