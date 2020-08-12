package featureflag

type MockClient struct {
	BoolVars   map[string]bool
	StringVars map[string]string
}

var _ Client = MockClient{}

func (c MockClient) Enabled(key, _ string) bool {
	return c.BoolVars[key]
}

func (c MockClient) Variation(key, defaultVal, _ string) string {
	res, ok := c.StringVars[key]
	if !ok {
		return defaultVal
	}
	return res
}
