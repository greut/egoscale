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

func (s *computeSecurityGroupTestSuite) TestSecurityGroupIngressRules() {
	res1, err := s.client.c.ListWithContext(s.client.ctx, &egoapi.SecurityGroup{Name: "default"})
	if err != nil {
		s.FailNow("unable to retrieve the default Security Group", err)
	}
	securityGroupDefault := res1[0].(*egoapi.SecurityGroup)

	res2, teardown, err := securityGroupFixture("", "")
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}
	defer teardown()

	securityGroup := s.client.securityGroupFromAPI(res2)

	for _, rule := range []*egoapi.AuthorizeSecurityGroupIngress{
		&egoapi.AuthorizeSecurityGroupIngress{
			SecurityGroupName: securityGroup.Name,
			Description:       "test-egoscale",
			CIDRList:          []egoapi.CIDR{*egoapi.MustParseCIDR("0.0.0.0/0")},
			StartPort:         8000,
			EndPort:           9000,
			Protocol:          "tcp",
		},
		&egoapi.AuthorizeSecurityGroupIngress{
			SecurityGroupName:     securityGroup.Name,
			Description:           "test-egoscale",
			UserSecurityGroupList: []egoapi.UserSecurityGroup{securityGroupDefault.UserSecurityGroup()},
			Protocol:              "icmp",
			IcmpType:              8,
			IcmpCode:              0,
		},
	} {
		if _, err := s.client.c.RequestWithContext(s.client.ctx, rule); err != nil {
			s.FailNow("unable to add a test rule to the fixture Security group", err)
		}
	}

	rules, err := securityGroup.IngressRules()
	if err != nil {
		s.FailNow("Security Group ingress rules listing failed", err)
	}
	assert.Len(s.T(), rules, 2)
	assert.NotEmpty(s.T(), rules[0].ID)
	assert.Equal(s.T(), "ingress", rules[0].Type)
	assert.Equal(s.T(), "test-egoscale", rules[0].Description)
	assert.Equal(s.T(), "default", rules[0].SecurityGroup.Name)
	assert.Equal(s.T(), "icmp", rules[0].Protocol)
	assert.Equal(s.T(), uint8(8), rules[0].ICMPType)
	assert.Equal(s.T(), uint8(0), rules[0].ICMPCode)
	assert.Equal(s.T(), "0.0.0.0/0", rules[1].NetworkCIDR.String())
	assert.Equal(s.T(), "tcp", rules[1].Protocol)
	assert.Equal(s.T(), "8000-9000", rules[1].Port)
}

func (s *computeSecurityGroupTestSuite) TestSecurityGroupEgressRules() {
	res, teardown, err := securityGroupFixture("", "")
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}
	defer teardown()

	securityGroup := s.client.securityGroupFromAPI(res)

	for _, rule := range []*egoapi.AuthorizeSecurityGroupEgress{
		&egoapi.AuthorizeSecurityGroupEgress{
			SecurityGroupName: securityGroup.Name,
			Description:       "DNS",
			CIDRList:          []egoapi.CIDR{*egoapi.MustParseCIDR("0.0.0.0/0")},
			StartPort:         53,
			EndPort:           53,
			Protocol:          "tcp",
		},
		&egoapi.AuthorizeSecurityGroupEgress{
			SecurityGroupName: securityGroup.Name,
			Description:       "DNS",
			CIDRList:          []egoapi.CIDR{*egoapi.MustParseCIDR("0.0.0.0/0")},
			StartPort:         53,
			EndPort:           53,
			Protocol:          "udp",
		},
	} {
		if _, err := s.client.c.RequestWithContext(s.client.ctx, rule); err != nil {
			s.FailNow("unable to add a test rule to the fixture Security group", err)
		}
	}

	rules, err := securityGroup.EgressRules()
	if err != nil {
		s.FailNow("Security Group egress rules listing failed", err)
	}
	assert.Len(s.T(), rules, 2)
	assert.NotEmpty(s.T(), rules[0].ID)
	assert.Equal(s.T(), "egress", rules[0].Type)
	assert.Equal(s.T(), "DNS", rules[0].Description)
	assert.Equal(s.T(), "0.0.0.0/0", rules[0].NetworkCIDR.String())
	assert.Equal(s.T(), "53", rules[0].Port)
	assert.Equal(s.T(), "tcp", rules[0].Protocol)
	assert.Equal(s.T(), "udp", rules[1].Protocol)
}

func (s *computeSecurityGroupTestSuite) TestSecurityGroupDelete() {
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

func (s *computeSecurityGroupTestSuite) TestSecurityGroupRuleDelete() {
	res1, teardown, err := securityGroupFixture("", "")
	if err != nil {
		s.FailNow("Security Group fixture setup failed", err)
	}
	defer teardown()

	securityGroup := s.client.securityGroupFromAPI(res1)

	res2, err := s.client.c.RequestWithContext(s.client.ctx, &egoapi.AuthorizeSecurityGroupIngress{
		SecurityGroupName: securityGroup.Name,
		CIDRList:          []egoapi.CIDR{*egoapi.MustParseCIDR("0.0.0.0/0")},
		StartPort:         22,
		EndPort:           22,
		Protocol:          "tcp",
	})
	if err != nil {
		s.FailNow("unable to add a test rule to the fixture Security group", err)
	}

	rule, err := s.client.securityGroupRuleFromAPI(&(res2.(*egoapi.SecurityGroup).IngressRule[0]))
	if err != nil {
		s.FailNow("Security Group rule retrieval failed", err)
	}

	if err := rule.Delete(); err != nil {
		s.FailNow("Security Group rule deletion failed", err)
	}

	res3, err := s.client.c.ListWithContext(s.client.ctx, &egoapi.SecurityGroup{Name: securityGroup.Name})
	if err != nil {
		s.FailNow("Security Group rules listing failed", err)
	}
	for _, item := range res3 {
		sg := item.(*egoapi.SecurityGroup)
		assert.Len(s.T(), sg.IngressRule, 0)
	}
}

func TestAccComputeSecurityGroupTestSuite(t *testing.T) {
	suite.Run(t, new(computeSecurityGroupTestSuite))
}
