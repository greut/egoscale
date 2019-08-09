package egoscale

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
	dir string
}

func (s *configTestSuite) SetupTest() {
	s.dir = os.TempDir()
}

func (s *configTestSuite) TearDownSuite() {
	os.RemoveAll(s.dir)
}

func (s *configTestSuite) TestLoadConfig() {
	var file = path.Join(s.dir, "config.toml")

	assert.Empty(s.T(), testConfigFileFixture(file, fmt.Sprintf(`
[[profiles]]
name = "alice"
api_key = "%s"
api_secret = "%s"

[[profiles]]
name = "bob"
api_key = "%s"
api_secret = "%s"
`,
		testAliceAPIKey,
		testAliceAPISecret,
		testBobAPIKey,
		testBobAPISecret)))

	config, err := loadConfig(file)
	assert.Empty(s.T(), err)
	assert.Len(s.T(), config.Profiles, 2)
	assert.Equal(s.T(), []ConfigProfile{
		{Name: "alice", APIKey: testAliceAPIKey, APISecret: testAliceAPISecret},
		{Name: "bob", APIKey: testBobAPIKey, APISecret: testBobAPISecret},
	}, config.Profiles)
}

func (s *configTestSuite) TestConfigGetProfile() {
	var file = path.Join(s.dir, "config.toml")

	assert.Empty(s.T(), testConfigFileFixture(file, fmt.Sprintf(`
[[profiles]]
name = "alice"
api_key = "%s"
api_secret = "%s"

[[profiles]]
name = "bob"
api_key = "%s"
api_secret = "%s"
`,
		testAliceAPIKey,
		testAliceAPISecret,
		testBobAPIKey,
		testBobAPISecret)))

	config, err := loadConfig(file)
	assert.Empty(s.T(), err)

	profile, err := config.getProfile("")
	assert.Empty(s.T(), err)
	assert.Equal(s.T(), &ConfigProfile{
		Name:      "alice",
		APIKey:    testAliceAPIKey,
		APISecret: testAliceAPISecret,
	}, profile)

	profile, err = config.getProfile("bob")
	assert.Empty(s.T(), err)
	assert.Equal(s.T(), &ConfigProfile{
		Name:      "bob",
		APIKey:    testBobAPIKey,
		APISecret: testBobAPISecret,
	}, profile)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}
