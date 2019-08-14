// +build testacc

package dns

import (
	"testing"

	egoerr "github.com/exoscale/egoscale/error"
	egoapi "github.com/exoscale/egoscale/internal/egoscale"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type dnsDomainTestSuite struct {
	suite.Suite
	client *Client

	testDomainName string
}

func (s *dnsDomainTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testDomainName = "example.net"
}

func (s *dnsDomainTestSuite) TestDomainAddRecord() {
	var (
		recordName     = "test-egoscale"
		recordType     = "MX"
		recordContent  = "mx1.example.net"
		recordPriority = 10
		recordTTL      = 1042
	)

	res, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck
	domain := s.client.domainFromAPI(res)

	record, err := domain.AddRecord(
		recordName,
		recordType,
		recordContent,
		recordPriority,
		recordTTL,
	)
	if err != nil {
		s.FailNow("domain record creation failed", err)
	}
	assert.Equal(s.T(), recordName, record.Name)
	assert.Equal(s.T(), recordType, record.Type)
	assert.Equal(s.T(), recordContent, record.Content)
	// assert.Equal(s.T(), recordPriority, record.Priority) // TODO: API bug, uncomment once fixed
	assert.Equal(s.T(), recordTTL, record.TTL)
}

func (s *dnsDomainTestSuite) TestDomainRecords() {
	var (
		recordName     = "test-egoscale"
		recordType     = "MX"
		recordContent  = "mx1.example.net"
		recordPriority = 10
		recordTTL      = 1042
	)

	res, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	if _, err = s.client.c.Request(&egoapi.CreateDNSRecord{
		Domain:   res.Name,
		Name:     recordName,
		Type:     recordType,
		Content:  recordContent,
		Priority: recordPriority,
		TTL:      recordTTL,
	}); err != nil {
		s.FailNow("domain record fixture setup failed", err)
	}
	domain := s.client.domainFromAPI(res)

	records, err := domain.Records()
	if err != nil {
		s.FailNow("domain records listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(records), 1)

	for _, record := range records {
		if record.Name == "" {
			continue
		}

		assert.Equal(s.T(), recordName, record.Name)
		assert.Equal(s.T(), recordType, record.Type)
		assert.Equal(s.T(), recordContent, record.Content)
		// assert.Equal(s.T(), recordPriority, record.Priority) // TODO: API bug, uncomment once fixed
		assert.Equal(s.T(), recordTTL, record.TTL)
	}
}

func (s *dnsDomainTestSuite) TestCreateDomain() {
	var (
		unicodeName          = "égzoskèle.ch"
		unicodeNamePunycoded = "xn--gzoskle-6xad.ch"
	)

	domain, err := s.client.CreateDomain(s.testDomainName)
	if err != nil {
		s.FailNow("domain creation failed", err)
	}
	assert.Greater(s.T(), domain.ID, int64(0))
	assert.Equal(s.T(), s.testDomainName, domain.Name)

	if _, err = s.client.c.Request(&egoapi.DeleteDNSDomain{Name: domain.Name}); err != nil {
		s.FailNow("domain deletion failed", err)
	}

	// With Unicode domain name
	domain, err = s.client.CreateDomain(unicodeName)
	if err != nil {
		s.FailNow("Unicode domain creation failed", err)
	}
	assert.Greater(s.T(), domain.ID, int64(0))
	assert.Equal(s.T(), unicodeNamePunycoded, domain.Name)
	assert.Equal(s.T(), unicodeName, domain.UnicodeName)

	if _, err = s.client.c.Request(&egoapi.DeleteDNSDomain{Name: domain.Name}); err != nil {
		s.FailNow("domain deletion failed", err)
	}
}

func (s *dnsDomainTestSuite) TestListDomains() {
	_, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	// We cannot guarantee that there will be only our resources,
	// so we ensure we get at least our fixture domain
	domains, err := s.client.ListDomains()
	if err != nil {
		s.FailNow("domains listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(domains), 1)
}

func (s *dnsDomainTestSuite) TestGetDomainByID() {
	res, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	domain, err := s.client.GetDomainByID(res.ID)
	if err != nil {
		s.FailNow("domain retrieval by ID failed", err)
	}
	assert.Equal(s.T(), s.testDomainName, domain.Name)

	domain, err = s.client.GetDomainByID(1)
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), domain)
}

func (s *dnsDomainTestSuite) TestGetDomainByName() {
	_, teardown, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	domain, err := s.client.GetDomainByName(s.testDomainName)
	if err != nil {
		s.FailNow("domain retrieval by name failed", err)
	}
	assert.Equal(s.T(), s.testDomainName, domain.Name)

	domain, err = s.client.GetDomainByName("lolnope")
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), domain)
}

func (s *dnsDomainTestSuite) TestDeleteDomain() {
	res, _, err := domainFixture(s.testDomainName)
	if err != nil {
		s.FailNow("domain fixture setup failed", err)
	}

	domain := s.client.domainFromAPI(res)
	domainName := domain.Name
	if err = domain.Delete(); err != nil {
		s.FailNow("domain deletion failed", err)
	}
	assert.Equal(s.T(), int64(0), domain.ID)
	assert.Empty(s.T(), domain.Name)

	// We have to list all domains and check if our test domain isn't in the
	// results since there is no way to search for a specific domain via the API
	domains, err := s.client.c.ListWithContext(s.client.ctx, &egoapi.DNSDomain{})
	if err != nil {
		s.FailNow("domains listing failed", err)
	}
	for _, d := range domains {
		assert.NotEqualf(s.T(), d.(*egoapi.DNSDomain).Name, domainName, "domain %q not deleted", domainName)
	}
}

func TestAccDNSDomainTestSuite(t *testing.T) {
	suite.Run(t, new(dnsDomainTestSuite))
}
