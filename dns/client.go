package dns

import (
	"context"
	"errors"

	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
)

const DefaultAPIEndpoint = "https://api.exoscale.com/compute"

// Client represents an Exoscale DNS API client.
type Client struct {
	c       *egoapi.Client
	ctx     context.Context
	tracing bool
}

func (c *Client) csError(err error) error {
	if _, ok := err.(*egoapi.ErrorResponse); ok {
		return errors.New(err.(*egoapi.ErrorResponse).ErrorText)
	}

	return err
}

// NewClient returns a new Exoscale DNS API client, configured to use apiKey and apiSecret as API credentials and
// apiEndpoint as an alternative DNS API endpoint URL. If tracing is true, the outgoing API calls and received
// responses will be displayed on the process standard error output.
func NewClient(ctx context.Context, apiKey, apiSecret, apiEndpoint string, tracing bool) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, egoerr.ErrMissingAPICredentials
	}

	if apiEndpoint == "" {
		apiEndpoint = DefaultAPIEndpoint
	}

	return &Client{
		c:       egoapi.NewClient(apiEndpoint, apiKey, apiSecret),
		ctx:     ctx,
		tracing: tracing,
	}, nil
}
