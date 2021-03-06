// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package testing

import (
	"os"
	"strings"

	"github.com/juju/loggo"
	"github.com/juju/testing"
	"github.com/juju/utils"
	"github.com/juju/utils/featureflag"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/juju/osenv"
	"github.com/juju/juju/network"
	"github.com/juju/juju/wrench"
)

var logger = loggo.GetLogger("juju.testing")

// JujuOSEnvSuite isolates the tests from Juju environment variables.
// This is intended to be only used by existing suites, usually embedded in
// BaseSuite and in FakeJujuHomeSuite. Eventually the tests relying on
// JujuOSEnvSuite will be converted to use the IsolationSuite in
// github.com/juju/testing, and this suite will be removed.
// Do not use JujuOSEnvSuite when writing new tests.
type JujuOSEnvSuite struct {
	oldJujuHome    string
	oldHomeEnv     string
	oldEnvironment map[string]string
}

func (s *JujuOSEnvSuite) SetUpSuite(c *gc.C) {
}

func (s *JujuOSEnvSuite) TearDownSuite(c *gc.C) {
}

func (s *JujuOSEnvSuite) SetUpTest(c *gc.C) {
	s.oldEnvironment = make(map[string]string)
	for _, name := range []string{
		osenv.JujuHomeEnvKey,
		osenv.JujuEnvEnvKey,
		osenv.JujuLoggingConfigEnvKey,
		osenv.JujuFeatureFlagEnvKey,
	} {
		s.oldEnvironment[name] = os.Getenv(name)
		os.Setenv(name, "")
	}
	s.oldHomeEnv = utils.Home()
	utils.SetHome("")
	// Update the feature flag set to be empty (given we have just set the
	// environment value to the empty string)
	featureflag.SetFlagsFromEnvironment(osenv.JujuFeatureFlagEnvKey)
	s.oldJujuHome = osenv.SetJujuHome("")
}

func (s *JujuOSEnvSuite) TearDownTest(c *gc.C) {
	for name, value := range s.oldEnvironment {
		os.Setenv(name, value)
	}
	utils.SetHome(s.oldHomeEnv)
	osenv.SetJujuHome(s.oldJujuHome)
}

func (s *JujuOSEnvSuite) SetFeatureFlags(flag ...string) {
	flags := strings.Join(flag, ",")
	if err := os.Setenv(osenv.JujuFeatureFlagEnvKey, flags); err != nil {
		panic(err)
	}
	logger.Debugf("setting feature flags: %s", flags)
	featureflag.SetFlagsFromEnvironment(osenv.JujuFeatureFlagEnvKey)
}

// BaseSuite provides required functionality for all test suites
// when embedded in a gocheck suite type:
// - logger redirect
// - no outgoing network access
// - protection of user's home directory
// - scrubbing of env vars
// TODO (frankban) 2014-06-09: switch to using IsolationSuite.
// NOTE: there will be many tests that fail when you try to change
// to the IsolationSuite that rely on external things in PATH.
type BaseSuite struct {
	testing.CleanupSuite
	testing.LoggingSuite
	JujuOSEnvSuite
}

func (s *BaseSuite) SetUpSuite(c *gc.C) {
	wrench.SetEnabled(false)
	s.CleanupSuite.SetUpSuite(c)
	s.LoggingSuite.SetUpSuite(c)
	s.JujuOSEnvSuite.SetUpSuite(c)
	s.PatchValue(&utils.OutgoingAccessAllowed, false)
}

func (s *BaseSuite) TearDownSuite(c *gc.C) {
	s.JujuOSEnvSuite.TearDownSuite(c)
	s.LoggingSuite.TearDownSuite(c)
	s.CleanupSuite.TearDownSuite(c)
}

func (s *BaseSuite) SetUpTest(c *gc.C) {
	s.CleanupSuite.SetUpTest(c)
	s.LoggingSuite.SetUpTest(c)
	s.JujuOSEnvSuite.SetUpTest(c)

	// We do this to isolate invocations of bash from pulling in the
	// ambient user environment, and potentially affecting the tests.
	// We can't always just use IsolationSuite because we still need
	// PATH and possibly a couple other envars.
	s.PatchEnvironment("BASH_ENV", "")
	network.ResetGobalPreferIPv6()
}

func (s *BaseSuite) TearDownTest(c *gc.C) {
	s.JujuOSEnvSuite.TearDownTest(c)
	s.LoggingSuite.TearDownTest(c)
	s.CleanupSuite.TearDownTest(c)
}
