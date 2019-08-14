package compute

import (
	"net"

	"github.com/exoscale/egoscale/api"
	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
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
	ICMPCode      int
	ICMPType      int

	c *Client
}

// TODO: SecurityGroupRule.Delete()

// func (c *Client) securityGroupRuleFromAPI(rule *egoapi.IngressRule)

// SecurityGroup represents a Security Group.
type SecurityGroup struct {
	api.Resource

	ID          string
	Name        string
	Description string

	c *Client
}

// TODO: SecurityGroup.Rules()

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
