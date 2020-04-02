package util

import "sync/atomic"

const (
	falseValue int32 = 0
	trueValue  int32 = 1
)

type AtomicBool interface {
	Set(bool) bool
	Get() bool
}

type atomicBool struct {
	value int32
}

func NewAtomicBool(val bool) AtomicBool {
	a := &atomicBool{value: falseValue}
	a.Set(val)
	return a
}

// Set will set the value to boolValue and will return the previous value
func (a *atomicBool) Set(boolValue bool) bool {
	intValue := int32(falseValue)
	if boolValue {
		intValue = trueValue
	}
	return toTruthy(atomic.SwapInt32(&a.value, intValue))
}

// Get will return the current value
func (a *atomicBool) Get() bool {
	return toTruthy(atomic.LoadInt32(&a.value))
}

func toTruthy(val int32) bool {
	return val != falseValue
}
