// +build testacc

package runstatus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type runstatusPageTestSuite struct {
	suite.Suite
	client *Client

	testPageName string
}

func (s *runstatusPageTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testPageName = "test-egoscale"
}

func (s *runstatusPageTestSuite) TestListPages() {
	_, teardown, err := pageFixture("")
	if err != nil {
		s.FailNow("page fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	// We cannot guarantee that there will be only our resources,
	// so we ensure we get at least our fixture page
	pages, err := s.client.ListPages()
	if err != nil {
		s.FailNow("pages listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(pages), 1)
}

func TestAccRunstatusPageTestSuite(t *testing.T) {
	suite.Run(t, new(runstatusPageTestSuite))
}
