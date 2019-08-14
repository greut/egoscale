// +build testacc

package compute

import (
	"testing"

	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type computeSecurityGroupTestSuite struct {
	suite.Suite
	client *Client

	testSecurityGroupName        string
	testSecurityGroupDescription string
}

func (s *computeSecurityGroupTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testSecurityGroupName = "test-egoscale"
	s.testSecurityGroupDescription = "Security Group created by the egoscale library"
}

func (s *computeSecurityGroupTestSuite) TestCreateSecurityGroup() {
	securityGroup, err := s.client.CreateSecurityGroup(s.testSecurityGroupName, s.testSecurityGroupDescription)
	if err != nil {
		s.FailNow("Security Group creation failed", err)
	}
	assert.NotEmpty(s.T(), securityGroup.ID)
	assert.Equal(s.T(), securityGroup.Name, s.testSecurityGroupName)
	assert.Equal(s.T(), securityGroup.Description, s.testSecurityGroupDescription)

	if _, err = s.client.c.Request(&egoapi.DeleteSecurityGroup{Name: securityGroup.Name}); err != nil {
		s.FailNow("Security Group deletion failed", err)
	}
}

func (s *computeSecurityGroupTestSuite) TestListSecurityGroups() {
	_, teardown, err := securityGroupFixture("", "")
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	// We cannot guarantee that there will be only our resources,
	// so we ensure we get at least our fixture SG + the default SG
	securityGroups, err := s.client.ListSecurityGroups()
	if err != nil {
		s.FailNow("Security Groups listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(securityGroups), 2)
}

func (s *computeSecurityGroupTestSuite) TestGetSecurityGroupByID() {
	res, teardown, err := securityGroupFixture(s.testSecurityGroupName, s.testSecurityGroupDescription)
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	securityGroup, err := s.client.GetSecurityGroupByID(res.ID.String())
	if err != nil {
		s.FailNow("Security Group retrieval by ID failed", err)
	}
	assert.Equal(s.T(), securityGroup.ID, res.ID.String())
	assert.Equal(s.T(), securityGroup.Name, res.Name)
	assert.Equal(s.T(), securityGroup.Description, res.Description)

	securityGroup, err = s.client.GetSecurityGroupByID("00000000-0000-0000-0000-000000000000")
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), securityGroup)
}

func (s *computeSecurityGroupTestSuite) TestGetSecurityGroupByName() {
	res, teardown, err := securityGroupFixture(s.testSecurityGroupName, s.testSecurityGroupDescription)
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	securityGroup, err := s.client.GetSecurityGroupByName(res.Name)
	if err != nil {
		s.FailNow("Security Group retrieval by name failed", err)
	}
	assert.Equal(s.T(), securityGroup.ID, res.ID.String())
	assert.Equal(s.T(), securityGroup.Name, res.Name)
	assert.Equal(s.T(), securityGroup.Description, res.Description)

	securityGroup, err = s.client.GetSecurityGroupByName("lolnope")
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), securityGroup)
}

func (s *computeSecurityGroupTestSuite) TestDeleteSecurityGroup() {
	res, _, err := securityGroupFixture("", "")
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}

	securityGroup := s.client.securityGroupFromAPI(res)
	securityGroupName := securityGroup.Name
	if err = securityGroup.Delete(); err != nil {
		s.FailNow("Security Group deletion failed", err)
	}
	assert.Empty(s.T(), securityGroup.ID)
	assert.Empty(s.T(), securityGroup.Name)
	assert.Empty(s.T(), securityGroup.Description)

	r, _ := s.client.c.ListWithContext(s.client.ctx, &egoapi.SecurityGroup{Name: securityGroupName})
	assert.Len(s.T(), r, 0)
}

func TestAccComputeSecurityGroupTestSuite(t *testing.T) {
	suite.Run(t, new(computeSecurityGroupTestSuite))
}
