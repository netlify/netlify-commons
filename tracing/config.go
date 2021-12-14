package tracing

import (
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	HeaderNFDebugLogging = "X-NF-Debug-Logging"
	HeaderRequestUUID    = "X-BB-CLIENT-REQUEST-UUID"
)

type Config struct {
	Enabled     bool   `default:"false"`
	Host        string `default:"localhost"`
	Port        string `default:"8126"`
	Tags        map[string]string
	EnableDebug bool `default:"false" split_words:"true" mapstructure:"enable_debug" json:"enable_debug" yaml:"enable_debug"`
}

func Configure(tc *Config, log logrus.FieldLogger, svcName string) {
	var t opentracing.Tracer = opentracing.NoopTracer{}
	if tc.Enabled {
		tracerAddr := fmt.Sprintf("%s:%s", tc.Host, tc.Port)
		tracerOps := []tracer.StartOption{
			tracer.WithService(svcName),
			tracer.WithAgentAddr(tracerAddr),
			tracer.WithDebugMode(tc.EnableDebug),
			tracer.WithLogger(debugLogger{log.WithField("component", "opentracing")}),
		}

		var serviceTagSet bool
		for k, v := range tc.Tags {
			if strings.ToLower(k) == "service" {
				serviceTagSet = true
			}

			tracerOps = append(tracerOps, tracer.WithGlobalTag(k, v))
		}

		if !serviceTagSet {
			tracerOps = append(tracerOps, tracer.WithGlobalTag("service", svcName))
		}

		t = opentracer.New(tracerOps...)
	}
	opentracing.SetGlobalTracer(t)
}

type debugLogger struct {
	log logrus.FieldLogger
}

func (l debugLogger) Log(msg string) {
	l.log.Debug(msg)
}
