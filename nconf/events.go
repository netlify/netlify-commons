package nconf

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ddstatsd "github.com/DataDog/datadog-go/statsd"
	"github.com/netlify/netlify-commons/metriks"
	"github.com/pkg/errors"
)

func sendDatadogEvents(conf metriks.Config, serviceName, version string) error {
	if !conf.Enabled {
		return nil
	}

	client, err := ddstatsd.New(conf.StatsdAddr())
	if err != nil {
		return errors.Wrap(err, "failed to connect to datadog agent")
	}

	tags := []string{
		fmt.Sprintf("version:%s", version),
		fmt.Sprintf("service:%s", serviceName),
	}
	host, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "failed to get the hostname")
	}

	key := "hostname"
	if val := os.Getenv("KUBERNETES_PORT"); val != "" {
		key = "pod"
	}
	tags = append(tags, fmt.Sprintf("%s:%s", key, host))

	start := &ddstatsd.Event{
		Tags:  append(tags, "event_type:startup"),
		Title: fmt.Sprintf("Service Start: %s", serviceName),
		Text:  fmt.Sprintf("Service %s @ %s is starting", serviceName, version),
	}
	if err := client.Event(start); err != nil {
		return errors.Wrap(err, "failed to send startup event")
	}

	done := &ddstatsd.Event{
		Tags:  append(tags, "event_type:shutdown"),
		Title: fmt.Sprintf("Service Shutdown: %s", serviceName),
		Text:  fmt.Sprintf("Service '%s @ %s' is stopping", serviceName, version),
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-signals
		_ = client.Event(done)
	}()

	return nil
}
