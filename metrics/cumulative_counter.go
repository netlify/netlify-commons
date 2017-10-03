package metrics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type CumulativeCounter interface {
	Increment(dims DimMap)
	IncrementN(val int64, dims DimMap)
}

type cumulativeCounter struct {
	metric
	mts      map[string]*metric
	statLock *sync.Mutex
}

func (e *Environment) NewCumulativeCounter(name string) CumulativeCounter {
	m := e.newMetric(name, CumulativeType, nil)

	c := &cumulativeCounter{
		metric:   *m,
		mts:      make(map[string]*metric),
		statLock: new(sync.Mutex),
	}

	e.reporter.register(c)
	return c
}

func (cc *cumulativeCounter) Increment(dims DimMap) {
	cc.IncrementN(1, dims)
}

func (cc *cumulativeCounter) IncrementN(val int64, dims DimMap) {
	cc.statLock.Lock()
	defer cc.statLock.Unlock()

	if dims == nil {
		dims = DimMap{}
	}
	sha, err := HashDims(dims)
	if err != nil {
		cc.env.reportError(&cc.RawMetric, errors.Wrap(err, "Failed to hash dimensions"))
		return
	}

	m, exists := cc.mts[sha]
	if !exists {
		// if we've never seen this before, we need to create a metric to track it
		m = cc.env.newMetric(cc.Name, CounterType, dims)
		cc.mts[sha] = m
	}

	m.Value += val
}

func (cc *cumulativeCounter) series() []*metric {
	res := []*metric{}
	for _, m := range cc.mts {
		res = append(res, m)
	}

	return res
}

func HashDims(dims DimMap) (string, error) {
	if dims == nil {
		return "", nil
	}

	data, err := json.Marshal(&dims)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
