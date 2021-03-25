package instrument

type Config struct {
	Key string `json:"key" yaml:"key"`
	// If this is false, instead of sending the event to Segment, only emits verbose log (for test/dev environment)
	Enabled bool `json:"enabled" yaml:"enabled" default:"false"`
}
