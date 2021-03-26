/*
Package instrument provides the tool to emit the events to the instrumentation
destination. Currently it supports Segment or logger as a destination.

In the config file, you can define the API key, as well as if it's enabled (use
Segment) or not (use logger).
	INSTRUMENT_ENABLED=true
	INSTRUMENT_KEY=segment_api_key

To use, you can import this package:
	import "github.com/netlify/netlify-commons/instrument"

You will likely need to import the Segment's analytics package as well, to
create new traits and properties.
	import "gopkg.in/segmentio/analytics-go.v3"

Then call the functions:
	instrument.Track("userid", "service:my_event", analytics.NewProperties().Set("color", "green"))

When enabling, make sure to follow the guideline specified in https://github.com/netlify/segment-events
*/
package instrument
