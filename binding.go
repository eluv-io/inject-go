package inject

import (
	"fmt"
	"reflect"
	"runtime"
)

type binding interface {
	fmt.Stringer
	// has to be a copy constructor
	// https://github.com/peter-edge/inject-go/commit/e525825afc80f0de819f35a6afc26a4bf3d3a192
	// this could be designed better
	resolvedBinding(*module, *injector) (resolvedBinding, error)
}

type resolvedBinding interface {
	fmt.Stringer
	validate() error
	get() (interface{}, error)
}

type intermediateBinding struct {
	bindingKey bindingKey
}

func newIntermediateBinding(to interface{}) binding {
	return &intermediateBinding{newBindingKey(reflect.TypeOf(to))}
}

func (i *intermediateBinding) String() string {
	return i.bindingKey.String()
}

func (i *intermediateBinding) resolvedBinding(module *module, injector *injector) (resolvedBinding, error) {
	binding, ok := module.binding(i.bindingKey)
	if !ok {
		return nil, errNoFinalBinding.withTag("bindingKey", i.bindingKey)
	}
	return binding.resolvedBinding(module, injector)
}

type singletonBinding struct {
	singleton interface{}
	injector  *injector
}

func newSingletonBinding(singleton interface{}) binding {
	return &singletonBinding{singleton, nil}
}

func (s *singletonBinding) String() string {
	return fmt.Sprintf("%v", s.singleton)
}

func (s *singletonBinding) validate() error {
	return nil
}

func (s *singletonBinding) get() (interface{}, error) {
	return s.singleton, nil
}

func (s *singletonBinding) resolvedBinding(module *module, injector *injector) (resolvedBinding, error) {
	return &singletonBinding{s.singleton, injector}, nil
}

type constructorBinding struct {
	constructor interface{}
	cache       *constructorBindingCache
	injector    *injector
}

type constructorBindingCache struct {
	numIn       int
	bindingKeys []bindingKey
}

func newConstructorBinding(constructor interface{}) binding {
	return &constructorBinding{constructor, newConstructorBindingCache(constructor), nil}
}

func newConstructorBindingCache(constructor interface{}) *constructorBindingCache {
	bindingKeys := getParameterBindingKeysForFunc(reflect.TypeOf(constructor))
	return &constructorBindingCache{len(bindingKeys), bindingKeys}
}

func (c *constructorBinding) String() string {
	return fmt.Sprintf("%v", c.constructor)
}

func (c *constructorBinding) validate() error {
	err := c.injector.validateBindingKeys(c.cache.bindingKeys)
	if err != nil {
		return unwrap(err).withTag("constructor", functionTag(c.constructor))
	}
	return nil
}

func (c *constructorBinding) get() (interface{}, error) {
	reflectValues, err := c.injector.getReflectValues(c.cache.bindingKeys)
	if err != nil {
		return nil, unwrap(err).withTag("constructor", functionTag(c.constructor))
	}
	return callConstructor(c.constructor, reflectValues)
}

func (c *constructorBinding) resolvedBinding(module *module, injector *injector) (resolvedBinding, error) {
	return &constructorBinding{c.constructor, c.cache, injector}, nil
}

type singletonConstructorBinding struct {
	constructorBinding
	loader *loader
}

func newSingletonConstructorBinding(constructor interface{}) binding {
	return &singletonConstructorBinding{constructorBinding{constructor, newConstructorBindingCache(constructor), nil}, nil}
}

func (s *singletonConstructorBinding) String() string {
	return fmt.Sprintf("%v", s.constructor)
}

func (s *singletonConstructorBinding) get() (interface{}, error) {
	return s.loader.load(s.constructorBinding.get)
}

func (s *singletonConstructorBinding) resolvedBinding(module *module, injector *injector) (resolvedBinding, error) {
	return &singletonConstructorBinding{constructorBinding{s.constructorBinding.constructor, s.constructorBinding.cache, injector}, newLoader()}, nil
}

type taggedConstructorBinding struct {
	constructor interface{}
	cache       *taggedConstructorBindingCache
	injector    *injector
}

type taggedConstructorBindingCache struct {
	inReflectType reflect.Type
	numFields     int
	bindingKeys   []bindingKey
}

func newTaggedConstructorBinding(constructor interface{}) binding {
	return &taggedConstructorBinding{constructor, newTaggedConstructorBindingCache(constructor), nil}
}

func newTaggedConstructorBindingCache(constructor interface{}) *taggedConstructorBindingCache {
	constructorReflectType := reflect.TypeOf(constructor)
	bindingKeys := getParameterBindingKeysForTaggedFunc(constructorReflectType)
	return &taggedConstructorBindingCache{constructorReflectType.In(0), len(bindingKeys), bindingKeys}
}

func (t *taggedConstructorBinding) String() string {
	return fmt.Sprintf("%v", t.constructor)
}

func (t *taggedConstructorBinding) validate() error {
	return t.injector.validateBindingKeys(t.cache.bindingKeys)
}

func (t *taggedConstructorBinding) get() (interface{}, error) {
	reflectValues, err := t.injector.getReflectValues(t.cache.bindingKeys)
	if err != nil {
		return nil, err
	}
	structReflectValue := newStructReflectValue(t.cache.inReflectType)
	populateStructReflectValue(&structReflectValue, reflectValues)
	return callConstructor(t.constructor, []reflect.Value{structReflectValue})
}

func (t *taggedConstructorBinding) resolvedBinding(module *module, injector *injector) (resolvedBinding, error) {
	return &taggedConstructorBinding{t.constructor, t.cache, injector}, nil
}

type taggedSingletonConstructorBinding struct {
	taggedConstructorBinding
	loader *loader
}

func newTaggedSingletonConstructorBinding(constructor interface{}) binding {
	return &taggedSingletonConstructorBinding{taggedConstructorBinding{constructor, newTaggedConstructorBindingCache(constructor), nil}, nil}
}

func (t *taggedSingletonConstructorBinding) String() string {
	return fmt.Sprintf("%v", t.constructor)
}

func (t *taggedSingletonConstructorBinding) get() (interface{}, error) {
	return t.loader.load(t.taggedConstructorBinding.get)
}

func (t *taggedSingletonConstructorBinding) resolvedBinding(module *module, injector *injector) (resolvedBinding, error) {
	return &taggedSingletonConstructorBinding{taggedConstructorBinding{t.taggedConstructorBinding.constructor, t.taggedConstructorBinding.cache, injector}, newLoader()}, nil
}

func callConstructor(constructor interface{}, reflectValues []reflect.Value) (interface{}, error) {
	returnValues := reflect.ValueOf(constructor).Call(reflectValues)
	if len(returnValues) == 2 {
		ret := returnValues[1].Interface()
		if ret != nil {
			return nil, errConstructorCall.withTag("err", ret, true).
				withTag("constructor", functionTag(constructor))
		}
	}
	return returnValues[0].Interface(), nil
}

func functionTag(fn interface{}) interface{} {
	return fmt.Sprintf("%s %T", runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), fn)
	// return reflect.TypeOf(fn)
}
