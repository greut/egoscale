// +build testacc

package dns

import (
	"testing"

	egoapi "github.com/exoscale/egoscale/internal/egoscale"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type dnsDomainRecordTestSuite struct {
	suite.Suite
	client *Client

	testDomainName string
}

func (s *dnsDomainRecordTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testDomainName = "example.net"
}

func (s *dnsDomainRecordTestSuite) TestDomainRecordUpdate() {
	var (
		recordName           = "test-egoscale"
		recordNameEdited     = "test-egoscale-edited"
		recordType           = "MX"
		recordContent        = "mx1.example.net"
		recordContentEdited  = "mx2.example.net"
		recordPriority       = 10
		recordPriorityEdited = 20
		recordTTL            = 1042
		recordTTLEdited      = 1043
	)

	domainRes, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck
	domain := s.client.domainFromAPI(domainRes)

	recordRes, err := s.client.c.Request(&egoapi.CreateDNSRecord{
		Domain:   domain.Name,
		Name:     recordName,
		Type:     recordType,
		Content:  recordContent,
		Priority: recordPriority,
		TTL:      recordTTL,
	})
	if err != nil {
		s.FailNow("domain record fixture creation failed", err)
	}
	record := s.client.domainRecordFromAPI(recordRes.(*egoapi.DNSRecord), domain)

	err = record.Update(recordNameEdited, recordContentEdited, recordPriorityEdited, recordTTLEdited)
	if err != nil {
		s.FailNow("domain record update failed", err)
	}
	assert.Equal(s.T(), recordNameEdited, record.Name)
	assert.Equal(s.T(), recordContentEdited, record.Content)
	// assert.Equal(s.T(), recordPriorityEdited, record.Priority) // TODO: API bug, uncomment once fixed
	assert.Equal(s.T(), recordTTLEdited, record.TTL)
}

func (s *dnsDomainRecordTestSuite) TestDomainRecordDelete() {
	var (
		recordName     = "test-egoscale"
		recordType     = "MX"
		recordContent  = "mx1.example.net"
		recordPriority = 10
		recordTTL      = 1042
	)

	domainRes, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck
	domain := s.client.domainFromAPI(domainRes)

	recordRes, err := s.client.c.Request(&egoapi.CreateDNSRecord{
		Domain:   domain.Name,
		Name:     recordName,
		Type:     recordType,
		Content:  recordContent,
		Priority: recordPriority,
		TTL:      recordTTL,
	})
	if err != nil {
		s.FailNow("domain record fixture creation failed", err)
	}
	record := s.client.domainRecordFromAPI(recordRes.(*egoapi.DNSRecord), domain)

	if err = record.Delete(); err != nil {
		s.FailNow("domain record deletion failed", err)
	}
	assert.Equal(s.T(), int64(0), record.ID)
	assert.Empty(s.T(), record.Name)
	assert.Empty(s.T(), record.Type)
	assert.Empty(s.T(), record.Content)
	assert.Equal(s.T(), int(0), record.Priority)
	assert.Equal(s.T(), int(0), record.TTL)
	assert.Empty(s.T(), record.Domain)

	// We have to list all records and check if our test record isn't in the
	// results since there is no way to search for a specific record via the API
	records, err := s.client.c.ListWithContext(s.client.ctx, &egoapi.ListDNSRecords{DomainID: domain.ID})
	if err != nil {
		s.FailNow("domain records listing failed", err)
	}
	for _, d := range records {
		assert.NotEqualf(s.T(), d.(*egoapi.DNSRecord).Name, recordName, "domain record %q not deleted", recordName)
	}
}

func TestAccDNSDomainRecordRecordTestSuite(t *testing.T) {
	suite.Run(t, new(dnsDomainRecordTestSuite))
}
