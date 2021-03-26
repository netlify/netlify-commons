package instrument

type Config struct {
	// Write Key for the Segment source
	Key string `json:"key" yaml:"key"`
	// If this is false, instead of sending the event to Segment, emits verbose log to logger
	Enabled bool `json:"enabled" yaml:"enabled" default:"false"`
}
