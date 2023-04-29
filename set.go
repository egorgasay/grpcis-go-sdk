package itisadb

import (
	"context"
	"errors"
	"fmt"
	"github.com/egorgasay/itisadb-go-sdk/api/balancer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	setDefault = -iota
	setToDB
	setToAll
	setToAllAndToDB
)

var ErrUniqueConstraint = errors.New("unique constraint failed")

func (c *Client) set(ctx context.Context, key, value string, server int32, uniques bool) (int32, error) {
	res, err := c.cl.Set(ctx, &balancer.BalancerSetRequest{
		Key:     key,
		Value:   value,
		Server:  server,
		Uniques: uniques,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			return 0, ErrUniqueConstraint
		}
		return 0, fmt.Errorf("an unknown error occurred while setting the value in the storage: %w", err)
	}

	return res.SavedTo, nil
}

// SetOne sets the value for the key to gRPCis.
func (c *Client) SetOne(ctx context.Context, key, value string, uniques bool) error {
	server, err := c.set(ctx, key, value, setDefault, uniques)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.keysAndServers[key] = server
	return nil
}

// SetTo sets the value for the key on the specified server.
func (c *Client) SetTo(ctx context.Context, key, value string, server int32, uniques bool) error {
	_, err := c.set(ctx, key, value, server, uniques)
	return err
}

// SetToDB sets the value for the key in the physical database.
func (c *Client) SetToDB(ctx context.Context, key, value string, uniques bool) error {
	_, err := c.set(ctx, key, value, setToDB, uniques)
	return err
}

// SetToAll sets the value for the key on all servers.
func (c *Client) SetToAll(ctx context.Context, key, value string, uniques bool) error {
	_, err := c.set(ctx, key, value, setToAll, uniques)
	return err
}

// SetToAllAndToDB sets the value for the key on all servers and in th physical database.
func (c *Client) SetToAllAndToDB(ctx context.Context, key, value string, uniques bool) error {
	_, err := c.set(ctx, key, value, setToAllAndToDB, uniques)
	return err
}

// SetMany sets a set of values for gRPCis.
func (c *Client) SetMany(ctx context.Context, keyValue map[string]string, uniques bool) error {
	for key, value := range keyValue {
		_, err := c.set(ctx, key, value, setDefault, uniques)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetManyOpts gets a lot of values from gRPCis with opts.
func (c *Client) SetManyOpts(ctx context.Context, keyValue map[string]Value, uniques bool) error {
	for key, value := range keyValue {
		_, err := c.set(ctx, key, value.Value, value.Opts.Server, uniques)
		if err != nil {
			return err
		}
	}
	return nil
}
