package user_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserModelSuite struct {
	suite.Suite
}

func (s *UserModelSuite) BeforeTest(suiteName, _ string) {
	user.TimeSince = func(t time.Time) time.Duration { return time.Date(2023, 6, 20, 0, 0, 0, 0, time.UTC).Sub(t) }
}

func (s *UserModelSuite) AfterTest(_, _ string) {
	user.TimeSince = time.Since
}

func (s *UserModelSuite) Test_IsAdmin() {
	a := assert.New(s.T())
	u := user.New("ragnar@lothbrok.com")
	u.IsAdmin = true

	b, _ := json.Marshal(u)
	a.Contains(string(b), "\"isAdmin\":true")

	// Now let's try with a false value
	u.IsAdmin = false

	b, _ = json.Marshal(u)

	// The isAdmin property should not be included in the JSON string.
	a.NotContains(string(b), "\"isAdmin\"")
}

func TestUser(t *testing.T) {
	suite.Run(t, &UserModelSuite{})
}
