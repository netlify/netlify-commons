package metriks

type Config struct {
	Host string `default:"localhost"`
	Port int    `default:"8125"`

	// Name is the typically the local hostname or pod name
	Name string `default:"local"`
	Tags map[string]string
}
