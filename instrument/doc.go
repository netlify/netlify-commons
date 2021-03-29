/*
Package instrument provides the tool to emit the events to the instrumentation
destination. Currently it supports Segment or logger as a destination.

When enabling, make sure to follow the guideline specified in https://github.com/netlify/segment-events

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

For testing, you can create your own mock instrument and use it:
	func TestSomething (t *testing.T) {
		log := testutil.TL(t)
		old := instrument.GetGlobalClient()
		t.Cleanup(func(){ instrument.SetGlobalClient(old) })
		instrument.SetGlobalClient(instrument.MockClient{Logger: log})
	}
*/
package instrument
