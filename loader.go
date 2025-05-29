package inject

import (
	"sync"
	"sync/atomic"

	"github.com/modern-go/gls"
)

// DetectCircularDependencies enables tracking of circular dependencies during the construction of dependencies.
var DetectCircularDependencies = false

type loader interface {
	load(binding resolvedBinding, f func() (interface{}, error)) (interface{}, error)
}

type valueErr struct {
	value interface{}
	err   error
}

type loaderImpl struct {
	once  sync.Once
	value atomic.Value
}

func newLoader() loader {
	if DetectCircularDependencies {
		return &traceLoader{}
	}
	return &loaderImpl{}
}

func (l *loaderImpl) load(_ resolvedBinding, f func() (interface{}, error)) (interface{}, error) {
	l.once.Do(func() {
		value, err := f()
		l.value.Store(&valueErr{value, err})
	})
	valueErr := l.value.Load().(*valueErr)
	return valueErr.value, valueErr.err
}

type traceLoader struct {
	value atomic.Value
	mutex sync.Mutex
	gid   atomic.Int64
}

func (l *traceLoader) load(binding resolvedBinding, f func() (interface{}, error)) (interface{}, error) {
	gid := gls.GoID()
	if !l.mutex.TryLock() {
		if gid == l.gid.Load() {
			return nil, errCircularDependency.
				withTag("call", " "+binding.String())
		}
		l.mutex.Lock()
	}
	l.gid.Store(gid)
	defer func() {
		l.mutex.Unlock()
		l.gid.Store(0)
	}()

	val := l.value.Load()
	if val == nil {
		value, err := f()
		val = &valueErr{value, err}
		l.value.Store(val)
	}
	ve := val.(*valueErr)
	return ve.value, ve.err
}
