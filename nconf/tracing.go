package nconf

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TracingConfig holds the options of the datadog opentracing host
type TracingConfig struct {
	Enabled bool   `default:"false"`
	Host    string `default:"localhost"`
	Port    string `default:"8126"`
	Tags    map[string]string
}

// ConfigureTracing sets a global tracer according to the config
func ConfigureTracing(config *TracingConfig, servName string) {
	var t opentracing.Tracer = opentracing.NoopTracer{}
	if config.Enabled {
		tracerAddr := fmt.Sprintf("%s:%s", config.Host, config.Port)
		tracerOps := []tracer.StartOption{
			tracer.WithServiceName(servName),
			tracer.WithAgentAddr(tracerAddr),
		}

		for k, v := range config.Tags {
			tracerOps = append(tracerOps, tracer.WithGlobalTag(k, v))
		}

		t = opentracer.New(tracerOps...)
	}
	opentracing.SetGlobalTracer(t)
}
