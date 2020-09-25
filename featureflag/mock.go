package featureflag

import (
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

type MockClient struct {
	BoolVars   map[string]bool
	StringVars map[string]string
}

var _ Client = MockClient{}

func (c MockClient) Enabled(key, userID string) bool {
	return c.EnabledUser(key, ld.NewUser(userID))
}

func (c MockClient) EnabledUser(key string, _ ld.User) bool {
	return c.BoolVars[key]
}

func (c MockClient) Variation(key string, defaultVal string, userID string) string {
	return c.VariationUser(key, defaultVal, ld.NewUser(userID))
}

func (c MockClient) VariationUser(key string, defaultVal string, _ ld.User) string {
	res, ok := c.StringVars[key]
	if !ok {
		return defaultVal
	}
	return res
}

func (c MockClient) AllEnabledFlags(string) []string {
	return []string{}
}
