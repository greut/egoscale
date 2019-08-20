package compute

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale/api"
	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
	"github.com/pkg/errors"
)

// SecurityGroupRule represents a Security Group rule.
type SecurityGroupRule struct {
	ID            string
	Type          string
	Description   string
	NetworkCIDR   *net.IPNet
	SecurityGroup *SecurityGroup
	Port          string
	Protocol      string
	ICMPType      uint8
	ICMPCode      uint8

	c *Client
}

// Delete deletes the Security Group rule.
func (r *SecurityGroupRule) Delete() error {
	var req egoapi.Command

	if r.Type == "ingress" {
		req = egoapi.RevokeSecurityGroupIngress{ID: egoapi.MustParseUUID(r.ID)}
	} else {
		req = egoapi.RevokeSecurityGroupEgress{ID: egoapi.MustParseUUID(r.ID)}
	}

	if err := r.c.csError(r.c.c.BooleanRequestWithContext(r.c.ctx, req)); err != nil {
		return err
	}

	r.ID = ""
	r.Type = ""
	r.Description = ""
	r.NetworkCIDR = nil
	r.SecurityGroup = nil
	r.Port = ""
	r.Protocol = ""
	r.ICMPType = 0
	r.ICMPCode = 0

	return nil
}

// func (r *SecurityGroupRule) parseRulePort() (uint16, uint16, error) {
// }

func (c *Client) securityGroupRuleFromAPI(v interface{}) (*SecurityGroupRule, error) {
	var (
		rule          *egoapi.IngressRule
		ruleType      string
		networkCIDR   *net.IPNet
		securityGroup *SecurityGroup
		port          string
		err           error
	)

	switch v.(type) {
	case *egoapi.IngressRule:
		ruleType = "ingress"
		rule = v.(*egoapi.IngressRule)

	case *egoapi.EgressRule:
		ruleType = "egress"
		rule = (*egoapi.IngressRule)(v.(*egoapi.EgressRule))
		// ^
		// Go typing madness: we cast the interface v underlying type *egoapi.EgressRule
		// into a *egoapi.IngressRule since the type egoapi.EgressRule is actually an alias for egoapi.IngressRule
		// Sorry about that...

	default:
		return nil, fmt.Errorf("invalid rule type from API: %T", v)
	}

	if rule.CIDR != nil {
		networkCIDR = &net.IPNet{IP: rule.CIDR.IP, Mask: rule.CIDR.Mask}
	}

	if rule.SecurityGroupName != "" {
		if securityGroup, err = c.GetSecurityGroupByName(rule.SecurityGroupName); err != nil {
			return nil, errors.Wrapf(err, "unable to retrieve Security Group %q", rule.SecurityGroupName)
		}
	}

	if rule.StartPort > 0 {
		if rule.StartPort < rule.EndPort {
			port = fmt.Sprintf("%d-%d", rule.StartPort, rule.EndPort)
		} else { // If StartPort is not lower than EndPort then it's equal since it can't be greater
			port = fmt.Sprint(rule.StartPort)
		}
	}

	return &SecurityGroupRule{
		ID:            rule.RuleID.String(),
		Type:          ruleType,
		Description:   rule.Description,
		NetworkCIDR:   networkCIDR,
		SecurityGroup: securityGroup,
		Port:          port,
		Protocol:      rule.Protocol,
		ICMPCode:      rule.IcmpCode,
		ICMPType:      rule.IcmpType,
		c:             c,
	}, nil
}

// SecurityGroup represents a Security Group.
type SecurityGroup struct {
	api.Resource

	ID          string
	Name        string
	Description string

	c *Client
}

// IngressRules returns the list of ingress-type Security Group rules.
func (s *SecurityGroup) IngressRules() ([]*SecurityGroupRule, error) {
	var (
		rules []*SecurityGroupRule
	)

	res, err := s.c.c.ListWithContext(s.c.ctx, &egoapi.SecurityGroup{Name: s.Name})
	if err != nil {
		return nil, err
	}

	for _, item := range res {
		sg := item.(*egoapi.SecurityGroup)

		rules = make([]*SecurityGroupRule, len(sg.IngressRule))
		for i, rule := range sg.IngressRule {
			if rules[i], err = s.c.securityGroupRuleFromAPI(&rule); err != nil {
				return nil, err
			}
		}
	}

	return rules, nil
}

// EgressRules returns the list of egress-type Security Group rules.
func (s *SecurityGroup) EgressRules() ([]*SecurityGroupRule, error) {
	var (
		rules []*SecurityGroupRule
	)

	res, err := s.c.c.ListWithContext(s.c.ctx, &egoapi.SecurityGroup{Name: s.Name})
	if err != nil {
		return nil, err
	}

	for _, item := range res {
		sg := item.(*egoapi.SecurityGroup)

		rules = make([]*SecurityGroupRule, len(sg.EgressRule))
		for i, rule := range sg.EgressRule {
			if rules[i], err = s.c.securityGroupRuleFromAPI(&rule); err != nil {
				return nil, err
			}
		}
	}

	return rules, nil
}

// TODO: SecurityGroup.AddRules()

// Delete deletes the Security Group.
func (sg *SecurityGroup) Delete() error {
	if err := sg.c.csError(sg.c.c.BooleanRequestWithContext(sg.c.ctx,
		&egoapi.DeleteSecurityGroup{Name: sg.Name})); err != nil {
		return err
	}

	sg.ID = ""
	sg.Name = ""
	sg.Description = ""

	return nil
}

// CreateSecurityGroup creates a new Security Group resource identified by name.
func (c *Client) CreateSecurityGroup(name, description string) (*SecurityGroup, error) {
	res, err := c.c.Request(&egoapi.CreateSecurityGroup{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}

	return c.securityGroupFromAPI(res.(*egoapi.SecurityGroup)), nil
}

// ListSecurityGroups returns the list of Security Groups.
func (c *Client) ListSecurityGroups() ([]*SecurityGroup, error) {
	res, err := c.c.ListWithContext(c.ctx, &egoapi.SecurityGroup{})
	if err != nil {
		return nil, err
	}

	securityGroups := make([]*SecurityGroup, 0)
	for _, i := range res {
		securityGroups = append(securityGroups, c.securityGroupFromAPI(i.(*egoapi.SecurityGroup)))
	}

	return securityGroups, nil
}

// GetSecurityGroupByName returns a Security Group by its name.
func (c *Client) GetSecurityGroupByName(name string) (*SecurityGroup, error) {
	return c.getSecurityGroup(nil, name)
}

// GetSecurityGroupByID returns a Security Group by its unique identifier.
func (c *Client) GetSecurityGroupByID(id string) (*SecurityGroup, error) {
	sgID, err := egoapi.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	return c.getSecurityGroup(sgID, "")
}

func (c *Client) getSecurityGroup(id *egoapi.UUID, name string) (*SecurityGroup, error) {
	res, err := c.c.ListWithContext(c.ctx, &egoapi.SecurityGroup{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, egoerr.ErrResourceNotFound
	}

	return c.securityGroupFromAPI(res[0].(*egoapi.SecurityGroup)), nil
}

func (c *Client) securityGroupFromAPI(sg *egoapi.SecurityGroup) *SecurityGroup {
	return &SecurityGroup{
		Resource:    api.MarshalResource(sg),
		ID:          sg.ID.String(),
		Name:        sg.Name,
		Description: sg.Description,
		c:           c,
	}
}
