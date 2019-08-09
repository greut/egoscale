package dns

import (
	"github.com/exoscale/egoscale/api"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
)

// DomainRecord represents a DNS domain record.
type DomainRecord struct {
	api.Resource

	ID       int64
	Type     string
	Name     string
	Content  string
	Priority int
	TTL      int
	Domain   *Domain

	c *Client
}

// Update updates the DNS domain record.
func (r *DomainRecord) Update(name, content string, priority, ttl int) error {
	cmd := egoapi.UpdateDNSRecord{
		DomainID: r.Domain.ID,
		ID:       r.ID,
		Type:     r.Type,
		Content:  r.Content,
		TTL:      r.TTL,
	}

	if name != "" {
		cmd.Name = name
	}
	if content != "" {
		cmd.Content = content
	}
	if priority > 0 {
		cmd.Priority = priority
	}
	if ttl > 0 {
		cmd.TTL = ttl
	}

	res, err := r.c.c.Request(&cmd)
	if err != nil {
		return err
	}
	record := res.(*egoapi.DNSRecord)

	r.Name = record.Name
	r.Content = record.Content
	r.Priority = record.Priority
	r.TTL = record.TTL

	return nil
}

// Delete deletes the DNS domain record.
func (r *DomainRecord) Delete() error {
	if err := r.c.csError(r.c.c.BooleanRequestWithContext(r.c.ctx, &egoapi.DeleteDNSRecord{
		DomainID: r.Domain.ID,
		ID:       r.ID,
	})); err != nil {
		return err
	}

	r.ID = 0
	r.Name = ""
	r.Type = ""
	r.Content = ""
	r.Priority = 0
	r.TTL = 0
	r.Domain = nil

	return nil
}

func (c *Client) domainRecordFromAPI(record *egoapi.DNSRecord, domain *Domain) *DomainRecord {
	return &DomainRecord{
		Resource: api.MarshalResource(record),
		ID:       record.ID,
		Type:     record.RecordType,
		Name:     record.Name,
		Content:  record.Content,
		Priority: record.Priority,
		TTL:      record.TTL,
		Domain:   domain,
		c:        c,
	}
}
