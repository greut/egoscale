package compute

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/exoscale/egoscale/internal/egoscale"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
)

func testRandomString() string {
	chars := "1234567890abcdefghijklmnopqrstuvwxyz"

	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 10)
	for i := range b {
		b[i] = chars[rand.Int63()%int64(len(chars))]
	}
	return string(b)
}

func testClientFromEnv() (*Client, error) {
	var (
		apiKey      = os.Getenv("EXOSCALE_API_KEY")
		apiSecret   = os.Getenv("EXOSCALE_API_SECRET")
		apiEndpoint = os.Getenv("EXOSCALE_COMPUTE_API_ENDPOINT")
	)

	return NewClient(context.Background(), apiKey, apiSecret, apiEndpoint, false)
}

func sshKeyFixture(name string) (*egoapi.SSHKeyPair, func() error, error) {
	if name == "" {
		name = "test-egoscale-" + testRandomString()
	}

	client, err := testClientFromEnv()
	if err != nil {
		return nil, nil, err
	}

	res, err := client.c.Request(&egoapi.CreateSSHKeyPair{Name: name})
	if err != nil {
		return nil, nil, err
	}

	return res.(*egoscale.SSHKeyPair),
		func() error {
			_, err := client.c.Request(&egoapi.DeleteSSHKeyPair{Name: name})
			return err
		},
		err
}
