// +build testacc

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type storageBucketTestSuite struct {
	suite.Suite
	client *Client

	testBucketName string
}

func (s *storageBucketTestSuite) SetupTest() {
	var err error

	if s.client, err = testClientFromEnv(); err != nil {
		s.FailNow("unable to initialize API client", err)
	}

	s.testBucketName = "test-egoscale"
}

// func (s *storageBucketTestSuite) TestCreateBucket() {
// }

func (s *storageBucketTestSuite) TestListBuckets() {
	_, teardown, err := bucketFixture("")
	if err != nil {
		s.FailNow("bucket fixture setup failed", err)
	}
	defer teardown() // nolint:errcheck

	// We cannot guarantee that there will be only our resources,
	// so we ensure we get at least our fixture bucket
	buckets, err := s.client.ListBuckets()
	if err != nil {
		s.FailNow("buckets listing failed", err)
	}
	assert.GreaterOrEqual(s.T(), len(buckets), 1)
}

// func (s *storageBucketTestSuite) TestGetBucket() {
// }

// func (s *storageBucketTestSuite) TestDeleteBucket() {
// }

func TestAccStorageBucketTestSuite(t *testing.T) {
	suite.Run(t, new(storageBucketTestSuite))
}
