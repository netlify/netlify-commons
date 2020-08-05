package metriks

type Config struct {
	Host string `default:"localhost"`
	Port int    `default:"8125"`

	Tags map[string]string

	// Deprecated: Name is not needed anymore.
	Name string
}
