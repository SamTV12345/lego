package ipv64

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey).WithDomain(envDomain)

func Test_splitDomain(t *testing.T) {
	testCases := []struct {
		desc       string
		domain     string
		prefix     string
		expected   string
		requireErr require.ErrorAssertionFunc
	}{
		{
			desc:       "empty",
			domain:     "",
			expected:   "",
			requireErr: require.Error,
		},
		{
			desc:       "missing sub domain",
			domain:     "home64.de",
			prefix:     "",
			expected:   "",
			requireErr: require.Error,
		},
		{
			desc:       "explicit domain: sub domain",
			domain:     "_acme-challenge.sub.home64.de",
			prefix:     "_acme-challenge",
			expected:   "sub.home64.de",
			requireErr: require.NoError,
		},
		{
			desc:       "explicit domain: subsub domain",
			domain:     "_acme-challenge.my.sub.home64.de",
			prefix:     "_acme-challenge.my",
			expected:   "sub.home64.de",
			requireErr: require.NoError,
		},
		{
			desc:       "explicit domain: subsubsub domain",
			domain:     "_acme-challenge.my.sub.sub.home64.de",
			prefix:     "_acme-challenge.my.sub",
			expected:   "sub.home64.de",
			requireErr: require.NoError,
		},
		{
			desc:       "only subname: sub domain",
			domain:     "_acme-challenge.sub",
			expected:   "",
			prefix:     "",
			requireErr: require.Error,
		},
		{
			desc:       "only subname: subsubsub domain",
			domain:     "_acme-challenge.my.sub.sub",
			expected:   "my.sub.sub",
			prefix:     "_acme-challenge",
			requireErr: require.NoError,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sub, root, err := splitDomain(test.domain)
			test.requireErr(t, err)

			assert.Equal(t, test.prefix, sub)
			assert.Equal(t, test.expected, root)
		})
	}
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "123",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "ipv64: some credentials information are missing: IPV64_API_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
		},
		{
			desc:     "missing credentials",
			expected: "ipv64: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
