// +build testacc

package compute

import (
	"testing"

	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type computeSSHKeyTestSuite struct {
	suite.Suite
	client *Client

	testSSHKeyName string
}

func (s *computeSSHKeyTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testSSHKeyName = "test-egoscale"
}

func (s *computeSSHKeyTestSuite) TestCreateSSHKey() {
	sshKey, err := s.client.CreateSSHKey(s.testSSHKeyName)
	if err != nil {
		s.FailNow("SSH key creation failed", err)
	}
	assert.Equal(s.T(), sshKey.Name, s.testSSHKeyName)
	assert.NotEmpty(s.T(), sshKey.Fingerprint)
	assert.NotEmpty(s.T(), sshKey.PrivateKey)

	if _, err = s.client.c.Request(&egoapi.DeleteSSHKeyPair{Name: sshKey.Name}); err != nil {
		s.FailNow("SSH key deletion failed", err)
	}
}

func (s *computeSSHKeyTestSuite) TestRegisterSSHKey() {
	var (
		publicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDGRYWaNYBG/Ld3ZnXGsK9pZ" +
			"l9kT3B6GXvsslgy/LCjkJvDIP+nL+opAArKZD1P1+SGylCLt8ISdJNNGLtxKp9CL12EGA" +
			"YqdDvm5PurkpqIkEsfhsIG4dne9hNu7ZW8aHGHDWM62/4uiWOKtbGdv/P33L/Fepzypwp" +
			"ivFsaXwPYVunAgoBQLUAmj/xcwtx7cvKS4zdj0+Iu21CIGU9wsH3ZLS34QiXtCGJyMOp1" +
			"58qld9Oeus3Y/7DQ4w5XvfGn9sddxHOSMwUlNiFVty673X3exgMIc8psZOsHvWZPS0zWx" +
			"9gEDE95cUU10K6u4vzTr2O6fgDOQBynEUw3CDiHvwRD alice@example.net"
		keyFingerprint = "a0:25:fa:32:c0:18:7a:f8:e8:b2:3b:30:d8:ca:9a:2e"
	)

	sshKey, err := s.client.RegisterSSHKey(s.testSSHKeyName, publicKey)
	if err != nil {
		s.FailNow("SSH key registration failed", err)
	}
	assert.Equal(s.T(), sshKey.Name, s.testSSHKeyName)
	assert.Equal(s.T(), sshKey.Fingerprint, keyFingerprint)

	if _, err = s.client.c.Request(&egoapi.DeleteSSHKeyPair{Name: sshKey.Name}); err != nil {
		s.FailNow("SSH key deletion failed", err)
	}
}

func (s *computeSSHKeyTestSuite) TestListSSHKeys() {
	_, teardown, err := sshKeyFixture("")
	if err != nil {
		s.FailNow("SSH key fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	// We cannot guarantee that there will be only our resources,
	// so we ensure we get at least our fixture SSH key
	sshKeys, err := s.client.ListSSHKeys()
	if err != nil {
		s.FailNow("SSH keys listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(sshKeys), 1)
}

func (s *computeSSHKeyTestSuite) TestGetSSHKey() {
	res, teardown, err := sshKeyFixture("")
	if err != nil {
		s.FailNow("SSH key fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	sshKey, err := s.client.GetSSHKey(res.Name)
	if err != nil {
		s.FailNow("SSH key retrieval failed", err)
	}
	assert.Equal(s.T(), sshKey.Name, res.Name)
	assert.NotEmpty(s.T(), sshKey.Fingerprint)

	sshKey, err = s.client.GetSSHKey("lolnope")
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), sshKey)
}

func (s *computeSSHKeyTestSuite) TesteSSHKeyDelete() {
	res, _, err := sshKeyFixture("")
	if err != nil {
		s.FailNow("SSH key fixture setup failed", err)
	}

	sshKey := s.client.sshKeyFromAPI(res)
	sshKeyName := sshKey.Name
	if err = sshKey.Delete(); err != nil {
		s.FailNow("SSH key deletion failed", err)
	}
	assert.Empty(s.T(), sshKey.Name)
	assert.Empty(s.T(), sshKey.Fingerprint)
	assert.Empty(s.T(), sshKey.PrivateKey)

	r, _ := s.client.c.ListWithContext(s.client.ctx, &egoapi.SSHKeyPair{Name: sshKeyName})
	assert.Len(s.T(), r, 0)
}

func TestAccComputeSSHKeyTestSuite(t *testing.T) {
	suite.Run(t, new(computeSSHKeyTestSuite))
}
