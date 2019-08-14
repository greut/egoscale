package storage

import (
	"context"
	"testing"

	egoerr "github.com/exoscale/egoscale/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type storageClientTestSuite struct {
	suite.Suite
}

func (s *storageClientTestSuite) TestNewClient() {
	client, err := NewClient(context.Background(), "", "", "", "", false)
	assert.EqualError(s.T(), err, egoerr.ErrMissingAPICredentials.Error())
	assert.Empty(s.T(), client)

	client, err = NewClient(context.Background(), "apiKey", "apiSecret", "apiEndpoint", "zone", false)
	if err != nil {
		s.FailNow("client instantiation failed", err)
	}
	assert.NotEmpty(s.T(), client)
}

func TestAccStorageClientTestSuite(t *testing.T) {
	suite.Run(t, new(storageClientTestSuite))
}
