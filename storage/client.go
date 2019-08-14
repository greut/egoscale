package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	egoerr "github.com/exoscale/egoscale/error"
)

// DefaultZone is the default Storage API zone.
const DefaultZone = "ch-gva-2"

// Client represents an Exoscale Storage API client.
type Client struct {
	s       *session.Session
	c       *s3.S3
	ctx     context.Context
	tracing bool
}

// NewClient returns a new Exoscale Storage API client, configured to use apiKey and apiSecret as API credentials,
// apiEndpoint as an alternative Storage API endpoint URL and zone as the storage zone. If tracing is true, the
// outgoing API calls and received responses will be displayed on the process standard error output.
func NewClient(ctx context.Context, apiKey, apiSecret, apiEndpoint, zone string, tracing bool) (*Client, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, egoerr.ErrMissingAPICredentials
	}

	if zone == "" {
		zone = DefaultZone
	}

	if apiEndpoint == "" {
		apiEndpoint = fmt.Sprintf("https://sos-%s.exo.io", zone)
	}

	sess, err := session.NewSessionWithOptions(session.Options{Config: aws.Config{
		Region:      aws.String(zone),
		Endpoint:    aws.String(apiEndpoint),
		Credentials: credentials.NewStaticCredentials(apiKey, apiSecret, ""),
		// TODO: support tracing using https://godoc.org/github.com/aws/aws-sdk-go/aws#Config
	}})
	if err != nil {
		return nil, err
	}

	return &Client{
		s:       sess,
		c:       s3.New(sess),
		ctx:     ctx,
		tracing: tracing,
	}, nil
}
