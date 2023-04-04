package inject

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
)

var injectorReflectType = reflect.TypeOf((*Injector)(nil))

type injector struct {
	// the injector's name
	name string
	// the parent injector for child injectors or nil otherwise
	parent *injector
	// resolved bindings
	bindings map[bindingKey]resolvedBinding
}

func newInjector(name string, modules ...Module) (*injector, error) {
	injector := &injector{
		name:     name,
		bindings: make(map[bindingKey]resolvedBinding),
	}
	return injector.init(modules)
}

func (inj *injector) init(modules []Module) (*injector, error) {
	modules = append(modules, inj.createInjectorModule())
	var eager []*singletonBuilder
	for _, m := range modules {
		castModule, ok := m.(*module)
		if !ok {
			return nil, errCannotCastModule
		}
		if err := inj.installModule(castModule); err != nil {
			return nil, err
		}
		eager = append(eager, castModule.eager...)
	}
	if err := inj.validate(newCtx(inj)); err != nil {
		return nil, err
	}

	for _, e := range eager {
		if e.t != nil {
			// create the singleton
			_, err := inj.get(newBindingKey(e.t))
			if err != nil {
				return nil, err
			}
		}
		if e.fn != nil {
			res, err := inj.Call(e.fn)
			if err != nil {
				return nil, err
			}
			if len(res) > 0 {
				if resErr, isErr := res[len(res)-1].(error); isErr {
					// the last return argument is a non-nil error - return that!
					return nil, resErr
				}
			}
		}
	}
	return inj, nil
}

func (inj *injector) createInjectorModule() Module {
	m := NewModule()
	m.Bind((*Injector)(nil)).ToSingleton(inj)
	return m
}

func (inj *injector) installModule(module *module) error {
	numBindingErrors := len(module.bindingErrors)
	if numBindingErrors > 0 {
		err := errBindingErrors
		for i := 0; i < numBindingErrors; i++ {
			err = err.withTag(strconv.Itoa(i+1), module.bindingErrors[i].Error())
		}
		return err
	}
	for bindingKey, binding := range module.bindings {
		if foundBinding, ok := inj.bindings[bindingKey]; ok {
			return errAlreadyBound.withTag("bindingKey", bindingKey).withTag("foundBinding", foundBinding)
		}
		// check parent bindings, but allow replacing the binding of the injector
		if inj.parent != nil && bindingKey.reflectType() != injectorReflectType {
			if foundBinding, ok := inj.parent.bindings[bindingKey]; ok {
				return errAlreadyBound.withTag("bindingKey", bindingKey).withTag("foundBinding", foundBinding).withTag("scope", "parent")
			}
		}
		resolvedBinding, err := binding.resolvedBinding(module, inj)
		if err != nil {
			return err
		}
		inj.bindings[bindingKey] = resolvedBinding
	}
	return nil
}

func (inj *injector) validate(ctx ctx) error {
	for key, resolvedBinding := range inj.bindings {
		if err := ctx.push(key, resolvedBinding); err != nil {
			return err
		}
		if err := resolvedBinding.validate(ctx); err != nil {
			return err
		}
		ctx.pop()
	}
	return nil
}

func (inj *injector) String() string {
	parent := ""
	if inj.parent != nil {
		parent = ", parent " + inj.parent.String()
	}
	return fmt.Sprintf("injector{%s}%s", inj.name, parent)
}

func (inj *injector) keyValueStrings() []string {
	strings := make([]string, len(inj.bindings))
	ii := 0
	for bindingKey, binding := range inj.bindings {
		var bindingString string
		if bindingKey.reflectType() == injectorReflectType {
			bindingString = fmt.Sprintf("this@%p", inj)
		} else {
			bindingString = binding.String()
		}
		strings[ii] = fmt.Sprintf("%s:%s", bindingKey.String(), bindingString)
		ii++
	}
	return strings
}

func (inj *injector) Get(from interface{}) (interface{}, error) {
	return inj.get(newBindingKey(reflect.TypeOf(from)))
}

func (inj *injector) DependencyTree() (DependencyTree, error) {
	c := newCtx(inj)
	err := inj.validate(c)
	if err != nil {
		return nil, err
	}
	return c.tree(), nil
}

func (inj *injector) GetTagged(tag string, from interface{}) (interface{}, error) {
	return inj.get(newTaggedBindingKey(reflect.TypeOf(from), tag))
}

func (inj *injector) GetTaggedBool(tag string) (bool, error) {
	obj, err := inj.getTaggedConstant(tag, boolConstantKind)
	if err != nil {
		return boolConstant, err
	}
	return obj.(bool), nil
}

func (inj *injector) GetTaggedInt(tag string) (int, error) {
	obj, err := inj.getTaggedConstant(tag, intConstantKind)
	if err != nil {
		return intConstant, err
	}
	return obj.(int), nil
}

func (inj *injector) GetTaggedInt8(tag string) (int8, error) {
	obj, err := inj.getTaggedConstant(tag, int8ConstantKind)
	if err != nil {
		return int8Constant, err
	}
	return obj.(int8), nil
}

func (inj *injector) GetTaggedInt16(tag string) (int16, error) {
	obj, err := inj.getTaggedConstant(tag, int16ConstantKind)
	if err != nil {
		return int16Constant, err
	}
	return obj.(int16), nil
}

func (inj *injector) GetTaggedInt32(tag string) (int32, error) {
	obj, err := inj.getTaggedConstant(tag, int32ConstantKind)
	if err != nil {
		return int32Constant, err
	}
	return obj.(int32), nil
}

func (inj *injector) GetTaggedInt64(tag string) (int64, error) {
	obj, err := inj.getTaggedConstant(tag, int64ConstantKind)
	if err != nil {
		return int64Constant, err
	}
	return obj.(int64), nil
}

func (inj *injector) GetTaggedUint(tag string) (uint, error) {
	obj, err := inj.getTaggedConstant(tag, uintConstantKind)
	if err != nil {
		return uintConstant, err
	}
	return obj.(uint), nil
}

func (inj *injector) GetTaggedUint8(tag string) (uint8, error) {
	obj, err := inj.getTaggedConstant(tag, uint8ConstantKind)
	if err != nil {
		return uint8Constant, err
	}
	return obj.(uint8), nil
}

func (inj *injector) GetTaggedUint16(tag string) (uint16, error) {
	obj, err := inj.getTaggedConstant(tag, uint16ConstantKind)
	if err != nil {
		return uint16Constant, err
	}
	return obj.(uint16), nil
}

func (inj *injector) GetTaggedUint32(tag string) (uint32, error) {
	obj, err := inj.getTaggedConstant(tag, uint32ConstantKind)
	if err != nil {
		return uint32Constant, err
	}
	return obj.(uint32), nil
}

func (inj *injector) GetTaggedUint64(tag string) (uint64, error) {
	obj, err := inj.getTaggedConstant(tag, uint64ConstantKind)
	if err != nil {
		return uint64Constant, err
	}
	return obj.(uint64), nil
}

func (inj *injector) GetTaggedFloat32(tag string) (float32, error) {
	obj, err := inj.getTaggedConstant(tag, float32ConstantKind)
	if err != nil {
		return float32Constant, err
	}
	return obj.(float32), nil
}

func (inj *injector) GetTaggedFloat64(tag string) (float64, error) {
	obj, err := inj.getTaggedConstant(tag, float64ConstantKind)
	if err != nil {
		return float64Constant, err
	}
	return obj.(float64), nil
}

func (inj *injector) GetTaggedComplex64(tag string) (complex64, error) {
	obj, err := inj.getTaggedConstant(tag, complex64ConstantKind)
	if err != nil {
		return complex64Constant, err
	}
	return obj.(complex64), nil
}

func (inj *injector) GetTaggedComplex128(tag string) (complex128, error) {
	obj, err := inj.getTaggedConstant(tag, complex128ConstantKind)
	if err != nil {
		return complex128Constant, err
	}
	return obj.(complex128), nil
}

func (inj *injector) GetTaggedString(tag string) (string, error) {
	obj, err := inj.getTaggedConstant(tag, stringConstantKind)
	if err != nil {
		return stringConstant, err
	}
	return obj.(string), nil
}

func (inj *injector) getTaggedConstant(tag string, constantKind constantKind) (interface{}, error) {
	return inj.get(newTaggedBindingKey(constantKind.reflectType(), tag))
}

func (inj *injector) Call(function interface{}) ([]interface{}, error) {
	funcReflectType := reflect.TypeOf(function)
	if err := verifyIsFunc(funcReflectType); err != nil {
		return nil, err
	}
	bindingKeys := getParameterBindingKeysForFunc(funcReflectType)
	if err := inj.validateBindingKeys(bindingKeys); err != nil {
		return nil, unwrap(err).withTag("funcReflectType", funcReflectType)
	}
	reflectValues, err := inj.getReflectValues(bindingKeys)
	if err != nil {
		return nil, unwrap(err).withTag("funcReflectType", funcReflectType)
	}
	returnValues := reflect.ValueOf(function).Call(reflectValues)
	return reflectValuesToValues(returnValues), nil
}

func (inj *injector) CallTagged(taggedFunction interface{}) ([]interface{}, error) {
	taggedFuncReflectType := reflect.TypeOf(taggedFunction)
	if err := verifyIsTaggedFunc(taggedFuncReflectType); err != nil {
		return nil, err
	}
	bindingKeys := getParameterBindingKeysForTaggedFunc(taggedFuncReflectType)
	if err := inj.validateBindingKeys(bindingKeys); err != nil {
		return nil, unwrap(err).withTag("funcReflectType", taggedFuncReflectType)
	}
	reflectValues, err := inj.getReflectValues(bindingKeys)
	if err != nil {
		return nil, unwrap(err).withTag("funcReflectType", taggedFuncReflectType)
	}
	structReflectValue := newStructReflectValue(taggedFuncReflectType.In(0))
	populateStructReflectValue(&structReflectValue, reflectValues)
	returnValues := reflect.ValueOf(taggedFunction).Call([]reflect.Value{structReflectValue})
	return reflectValuesToValues(returnValues), nil
}

func (inj *injector) Populate(populateStructPtr interface{}) error {
	populateStructPtrReflectType := reflect.TypeOf(populateStructPtr)
	if err := verifyIsStructPtr(populateStructPtrReflectType); err != nil {
		return err
	}
	populateStructValue := reflect.Indirect(reflect.ValueOf(populateStructPtr))
	if err := verifyStructCanBePopulated(populateStructValue.Type()); err != nil {
		return unwrap(err).withTag("funcReflectType", populateStructPtr)
	}
	bindingKeys := getStructFieldBindingKeys(populateStructValue.Type())
	if err := inj.validateBindingKeys(bindingKeys); err != nil {
		return unwrap(err).withTag("funcReflectType", populateStructPtr)
	}
	reflectValues, err := inj.getReflectValues(bindingKeys)
	if err != nil {
		return unwrap(err).withTag("funcReflectType", populateStructPtr)
	}
	populateStructReflectValue(&populateStructValue, reflectValues)
	return nil
}

func (inj *injector) NewChildInjector(overridesType interface{}, modules ...Module) (Injector, error) {
	name := callerName(3, "child")
	return inj.NewNamedChildInjector(name, overridesType, modules...)
}

func (inj *injector) NewNamedChildInjector(name string, overridesType interface{}, modules ...Module) (Injector, error) {
	overrides, _ := inj.Get(overridesType)
	if om, ok := overrides.(Module); ok {
		modules = []Module{Override(modules...).With(om)}
	}
	injector := &injector{
		name:     name,
		parent:   inj,
		bindings: make(map[bindingKey]resolvedBinding),
	}
	_, err := injector.init(modules)
	if err != nil {
		return nil, err
	}
	return injector, nil
}

func (inj *injector) get(bindingKey bindingKey) (interface{}, error) {
	binding, err := inj.getBinding(bindingKey)
	if err != nil {
		return nil, err
	}
	return binding.get()
}

func (inj *injector) getBinding(bindingKey bindingKey, nostack ...bool) (resolvedBinding, error) {
	// get binding from parent, if any, but not the injector itself
	if inj.parent != nil && bindingKey.reflectType() != injectorReflectType {
		binding, err := inj.parent.getBinding(bindingKey, true)
		if err == nil {
			return binding, nil
		}
	}
	// get local binding
	binding, ok := inj.bindings[bindingKey]
	if !ok {
		return nil, errNoBinding.withTag("bindingKey", bindingKey, nostack...)
	}
	return binding, nil
}

func (inj *injector) getReflectValues(bindingKeys []bindingKey) ([]reflect.Value, error) {
	numBindingKeys := len(bindingKeys)
	reflectValues := make([]reflect.Value, numBindingKeys)
	for ii := 0; ii < numBindingKeys; ii++ {
		value, err := inj.get(bindingKeys[ii])
		if err != nil {
			return nil, err
		}
		reflectValues[ii] = reflect.ValueOf(value)
	}
	return reflectValues, nil
}

func (inj *injector) validateBindingKeys(bindingKeys []bindingKey) error {
	for _, bindingKey := range bindingKeys {
		if _, err := inj.getBinding(bindingKey); err != nil {
			return err
		}
	}
	return nil
}

// validate all bindings for the given binding keys recursively
func (inj *injector) validateBindings(ctx ctx, bindingKeys []bindingKey) error {
	for _, bindingKey := range bindingKeys {
		resolvedBinding, err := inj.getBinding(bindingKey)
		if err != nil {
			return err
		}
		if err := ctx.push(bindingKey, resolvedBinding); err != nil {
			return err
		}
		if err := resolvedBinding.validate(ctx); err != nil {
			return err
		}
		ctx.pop()
	}
	return nil
}

func verifyIsStructPtr(reflectType reflect.Type) error {
	if !isStructPtr(reflectType) {
		return errNotStructPtr.withTag("reflectType", reflectType)
	}
	return nil
}

func reflectValuesToValues(reflectValues []reflect.Value) []interface{} {
	lenReflectValues := len(reflectValues)
	values := make([]interface{}, lenReflectValues)
	for i := 0; i < lenReflectValues; i++ {
		values[i] = reflectValues[i].Interface()
	}
	return values
}

func callerName(skip int, defName string) string {
	res := defName
	pcs := make([]uintptr, 1)
	runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs)
	frame, _ := frames.Next()
	if frame.Function != "" {
		res = fmt.Sprintf("%s:%d\t%s", frame.File, frame.Line, frame.Function)
	}
	return res
}
