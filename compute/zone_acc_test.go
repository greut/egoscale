// +build testacc

package compute

import (
	"testing"

	egoerr "github.com/exoscale/egoscale/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type computeZoneTestSuite struct {
	suite.Suite
	client *Client

	testZoneID   string
	testZoneName string
}

func (s *computeZoneTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testZoneID = "1128bd56-b4d9-4ac6-a7b9-c715b187ce11"
	s.testZoneName = "ch-gva-2"
}

func (s *computeZoneTestSuite) TestListZones() {
	var expectedZones = []string{
		"at-vie-1",
		"bg-sof-1",
		"ch-dk-2",
		"ch-gva-2",
		"de-fra-1",
		"de-muc-1",
	}

	zones, err := s.client.ListZones()
	if err != nil {
		s.FailNow("zones listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(zones), len(expectedZones))
}

func (s *computeZoneTestSuite) TestGetZoneByID() {
	zone, err := s.client.GetZoneByID(s.testZoneID)
	if err != nil {
		s.FailNow("zone retrieval by ID failed", err)
	}
	assert.Equal(s.T(), zone.ID, s.testZoneID)
	assert.Equal(s.T(), zone.Name, s.testZoneName)

	zone, err = s.client.GetZoneByID("00000000-0000-0000-0000-000000000000")
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), zone)
}

func (s *computeZoneTestSuite) TestGetZoneByName() {
	zone, err := s.client.GetZoneByName(s.testZoneName)
	if err != nil {
		s.FailNow("zone retrieval by name failed", err)
	}
	assert.Equal(s.T(), zone.ID, s.testZoneID)
	assert.Equal(s.T(), zone.Name, s.testZoneName)

	zone, err = s.client.GetZoneByName("lolnope")
	assert.EqualError(s.T(), err, egoerr.ErrResourceNotFound.Error())
	assert.Empty(s.T(), zone)
}

func TestAccComputeZoneTestSuite(t *testing.T) {
	suite.Run(t, new(computeZoneTestSuite))
}
