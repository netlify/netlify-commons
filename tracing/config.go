package tracing

import (
	"fmt"

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
	UseDatadog  bool `default:"false" split_words:"true" mapstructure:"use_datadog" json:"use_datadog" yaml:"use_datadog"`
}

func Configure(tc *Config, log logrus.FieldLogger, svcName string) {
	var t opentracing.Tracer = opentracing.NoopTracer{}
	if tc.Enabled {
		tracerAddr := fmt.Sprintf("%s:%s", tc.Host, tc.Port)
		tracerOps := []tracer.StartOption{
			tracer.WithServiceName(svcName),
			tracer.WithAgentAddr(tracerAddr),
			tracer.WithDebugMode(tc.EnableDebug),
			tracer.WithLogger(debugLogger{log.WithField("component", "opentracing")}),
		}

		for k, v := range tc.Tags {
			tracerOps = append(tracerOps, tracer.WithGlobalTag(k, v))
		}

		if tc.UseDatadog {
			tracer.Start(tracerOps...)
		} else {
			t = opentracer.New(tracerOps...)
		}
	}
	opentracing.SetGlobalTracer(t)
}

type debugLogger struct {
	log logrus.FieldLogger
}

func (l debugLogger) Log(msg string) {
	l.log.Debug(msg)
}
