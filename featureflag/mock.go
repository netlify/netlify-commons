package featureflag

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
)

type MockClient struct {
	BoolVars   map[string]bool
	StringVars map[string]string
	IntVars    map[string]int
}

var _ Client = MockClient{}

func (c MockClient) Enabled(key, userID string, _ ...Attr) bool {
	return c.EnabledUser(key, lduser.NewUser(userID))
}

func (c MockClient) EnabledUser(key string, _ lduser.User) bool {
	return c.BoolVars[key]
}

func (c MockClient) Variation(key string, defaultVal string, userID string, _ ...Attr) string {
	return c.VariationUser(key, defaultVal, lduser.NewUser(userID))
}

func (c MockClient) VariationUser(key string, defaultVal string, _ lduser.User) string {
	res, ok := c.StringVars[key]
	if !ok {
		return defaultVal
	}
	return res
}

func (c MockClient) Int(key string, defaultVal int, userID string, _ ...Attr) int {
	return c.IntUser(key, defaultVal, lduser.NewUser(userID))
}

func (c MockClient) IntUser(key string, defaultVal int, _ lduser.User) int {
	res, ok := c.IntVars[key]
	if !ok {
		return defaultVal
	}
	return res
}

func (c MockClient) AllEnabledFlags(key string) []string {
	return c.AllEnabledFlagsUser(key, lduser.NewUser(key))
}

func (c MockClient) AllEnabledFlagsUser(key string, _ lduser.User) []string {
	var res []string
	for key, value := range c.BoolVars {
		if value {
			res = append(res, key)
		}
	}
	return res
}
