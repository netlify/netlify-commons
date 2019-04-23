package tracing

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	HeaderNFDebugLogging = "X-NF-Debug-Logging"
	HeaderRequestUUID    = "X-BB-CLIENT-REQUEST-UUID"
)

type Config struct {
	Enabled bool   `default:"false"`
	Host    string `default:"localhost"`
	Port    string `default:"8126"`
	Tags    map[string]string
}

func Configure(tc *Config, svcName string) {
	var t opentracing.Tracer = opentracing.NoopTracer{}
	if tc.Enabled {
		tracerAddr := fmt.Sprintf("%s:%s", tc.Host, tc.Port)
		tracerOps := []tracer.StartOption{
			tracer.WithServiceName(svcName),
			tracer.WithAgentAddr(tracerAddr),
		}

		for k, v := range tc.Tags {
			tracerOps = append(tracerOps, tracer.WithGlobalTag(k, v))
		}

		t = opentracer.New(tracerOps...)
	}
	opentracing.SetGlobalTracer(t)
}
