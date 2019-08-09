package dns

import (
	"context"
	"os"

	"github.com/exoscale/egoscale/internal/egoscale"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
)

func testClientFromEnv() (*Client, error) {
	var (
		apiKey      = os.Getenv("EXOSCALE_API_KEY")
		apiSecret   = os.Getenv("EXOSCALE_API_SECRET")
		apiEndpoint = os.Getenv("EXOSCALE_DNS_API_ENDPOINT")
	)

	return NewClient(context.Background(), apiKey, apiSecret, apiEndpoint, false)
}

func domainFixture(name string) (*egoapi.DNSDomain, func() error, error) {
	client, err := testClientFromEnv()
	if err != nil {
		return nil, nil, err
	}

	res, err := client.c.Request(&egoapi.CreateDNSDomain{Name: name})
	if err != nil {
		return nil, nil, err
	}

	return res.(*egoscale.DNSDomain),
		func() error {
			_, err := client.c.Request(&egoapi.DeleteDNSDomain{Name: name})
			return err
		},
		err
}
