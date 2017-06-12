package metrics

type Event struct {
	Name string `json:"name"`
	Dims DimMap `json:"dimensions"`

	// nanoseconds since epoch
	Timestamp int64 `json:"timestamp"`

	// these aren't intended on being indexed
	Props DimMap `json:"properties"`

	env *Environment
}

func (e *Environment) NewEvent(name string, dims, props DimMap) *Event {
	return &Event{
		Name:  name,
		Dims:  dims,
		Props: props,
		env:   e,
	}
}

func (e *Event) Record() error {
	return e.env.sendEvent(e)
}
