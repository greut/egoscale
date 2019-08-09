package compute

import (
	"github.com/exoscale/egoscale/api"
	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
)

// SSHKey represents a SSH key resource.
type SSHKey struct {
	api.Resource

	Name        string
	Fingerprint string
	PrivateKey  string

	c *Client
}

// Delete deletes the SSH Key.
func (k *SSHKey) Delete() error {
	if err := k.c.csError(k.c.c.BooleanRequestWithContext(k.c.ctx,
		&egoapi.DeleteSSHKeyPair{Name: k.Name})); err != nil {
		return err
	}

	k.Name = ""
	k.Fingerprint = ""
	k.PrivateKey = ""

	return nil
}

// CreateSSHKey creates a new SSH key resource identified by name, and returns an SSHKey object containing the
// corresponding SSH private key if successful or an error.
func (c *Client) CreateSSHKey(name string) (*SSHKey, error) {
	res, err := c.c.Request(&egoapi.CreateSSHKeyPair{Name: name})
	if err != nil {
		return nil, err
	}

	return c.sshKeyFromAPI(res.(*egoapi.SSHKeyPair)), nil
}

// RegisterSSHKey registers an existing SSH public key as a new resource identified by name, and returns an SSHKey
// object if successful or an error.
func (c *Client) RegisterSSHKey(name, publicKey string) (*SSHKey, error) {
	res, err := c.c.Request(&egoapi.RegisterSSHKeyPair{
		Name:      name,
		PublicKey: publicKey,
	})
	if err != nil {
		return nil, err
	}

	return c.sshKeyFromAPI(res.(*egoapi.SSHKeyPair)), nil
}

// ListSSHKeys returns the list of available Exoscale zones, or an error if the API query failed.
func (c *Client) ListSSHKeys() ([]*SSHKey, error) {
	res, err := c.c.ListWithContext(c.ctx, &egoapi.SSHKeyPair{})
	if err != nil {
		return nil, err
	}

	sshKeys := make([]*SSHKey, 0)
	for _, i := range res {
		sshKey := i.(*egoapi.SSHKeyPair)
		sshKeys = append(sshKeys, &SSHKey{
			Resource:    api.MarshalResource(sshKey),
			Name:        sshKey.Name,
			Fingerprint: sshKey.Fingerprint,
			PrivateKey:  sshKey.PrivateKey,
			c:           c,
		})
	}

	return sshKeys, nil
}

// GetSSHKey returns a SSH key by its name.
func (c *Client) GetSSHKey(name string) (*SSHKey, error) {
	res, err := c.c.ListWithContext(c.ctx, &egoapi.SSHKeyPair{Name: name})
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, egoerr.ErrResourceNotFound
	}

	return c.sshKeyFromAPI(res[0].(*egoapi.SSHKeyPair)), nil
}

func (c *Client) sshKeyFromAPI(sshKey *egoapi.SSHKeyPair) *SSHKey {
	return &SSHKey{
		Resource:    api.MarshalResource(sshKey),
		Name:        sshKey.Name,
		Fingerprint: sshKey.Fingerprint,
		PrivateKey:  sshKey.PrivateKey,
		c:           c,
	}
}
