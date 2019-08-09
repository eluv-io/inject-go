package inject

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var injectorReflectType = reflect.TypeOf((*Injector)(nil))

type injector struct {
	// the parent injector for child injectors or nil otherwise
	parent *injector
	// resolved bindings
	bindings map[bindingKey]resolvedBinding
}

func newInjector(modules []Module) (Injector, error) {
	injector := &injector{nil, make(map[bindingKey]resolvedBinding)}
	return initInjector(injector, modules)
}

func initInjector(injector *injector, modules []Module) (Injector, error) {
	modules = append(modules, createInjectorModule(injector))
	var eager []*singletonBuilder
	for _, m := range modules {
		castModule, ok := m.(*module)
		if !ok {
			return nil, errCannotCastModule
		}
		if err := installModuleToInjector(injector, castModule); err != nil {
			return nil, err
		}
		eager = append(eager, castModule.eager...)
	}
	if err := validate(injector); err != nil {
		return nil, err
	}
	for _, e := range eager {
		// create the singleton
		_, err := injector.get(newBindingKey(e.t))
		if err != nil {
			return nil, err
		}
		if e.fn != nil {
			_, err := injector.Call(e.fn)
			if err != nil {
				return nil, err
			}
		}
	}
	return injector, nil
}

func createInjectorModule(injector *injector) Module {
	m := NewModule()
	m.Bind((*Injector)(nil)).ToSingleton(injector)
	return m
}

func installModuleToInjector(injector *injector, module *module) error {
	numBindingErrors := len(module.bindingErrors)
	if numBindingErrors > 0 {
		err := errBindingErrors
		for i := 0; i < numBindingErrors; i++ {
			err = err.withTag(strconv.Itoa(i+1), module.bindingErrors[i].Error())
		}
		return err
	}
	for bindingKey, binding := range module.bindings {
		if foundBinding, ok := injector.bindings[bindingKey]; ok {
			return errAlreadyBound.withTag("bindingKey", bindingKey).withTag("foundBinding", foundBinding)
		}
		// check parent bindings, but allow replacing the binding of the injector
		if injector.parent != nil && bindingKey.reflectType() != injectorReflectType {
			if foundBinding, ok := injector.parent.bindings[bindingKey]; ok {
				return errAlreadyBound.withTag("bindingKey", bindingKey).withTag("foundBinding", foundBinding).withTag("scope", "parent")
			}
		}
		resolvedBinding, err := binding.resolvedBinding(module, injector)
		if err != nil {
			return err
		}
		injector.bindings[bindingKey] = resolvedBinding
	}
	return nil
}

func validate(injector *injector) error {
	for _, resolvedBinding := range injector.bindings {
		if err := resolvedBinding.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (i *injector) String() string {
	parent := ""
	if i.parent != nil {
		parent = ", parent " + i.parent.String()
	}
	return fmt.Sprintf("injector{%s}%s", strings.Join(i.keyValueStrings(), " "), parent)
}

func (i *injector) keyValueStrings() []string {
	strings := make([]string, len(i.bindings))
	ii := 0
	for bindingKey, binding := range i.bindings {
		var bindingString string
		if bindingKey.reflectType() == injectorReflectType {
			bindingString = fmt.Sprintf("this@%p", i)
		} else {
			bindingString = binding.String()
		}
		strings[ii] = fmt.Sprintf("%s:%s", bindingKey.String(), bindingString)
		ii++
	}
	return strings
}

func (i *injector) Get(from interface{}) (interface{}, error) {
	return i.get(newBindingKey(reflect.TypeOf(from)))
}

func (i *injector) GetTagged(tag string, from interface{}) (interface{}, error) {
	return i.get(newTaggedBindingKey(reflect.TypeOf(from), tag))
}

func (i *injector) GetTaggedBool(tag string) (bool, error) {
	obj, err := i.getTaggedConstant(tag, boolConstantKind)
	if err != nil {
		return boolConstant, err
	}
	return obj.(bool), nil
}

func (i *injector) GetTaggedInt(tag string) (int, error) {
	obj, err := i.getTaggedConstant(tag, intConstantKind)
	if err != nil {
		return intConstant, err
	}
	return obj.(int), nil
}

func (i *injector) GetTaggedInt8(tag string) (int8, error) {
	obj, err := i.getTaggedConstant(tag, int8ConstantKind)
	if err != nil {
		return int8Constant, err
	}
	return obj.(int8), nil
}

func (i *injector) GetTaggedInt16(tag string) (int16, error) {
	obj, err := i.getTaggedConstant(tag, int16ConstantKind)
	if err != nil {
		return int16Constant, err
	}
	return obj.(int16), nil
}

func (i *injector) GetTaggedInt32(tag string) (int32, error) {
	obj, err := i.getTaggedConstant(tag, int32ConstantKind)
	if err != nil {
		return int32Constant, err
	}
	return obj.(int32), nil
}

func (i *injector) GetTaggedInt64(tag string) (int64, error) {
	obj, err := i.getTaggedConstant(tag, int64ConstantKind)
	if err != nil {
		return int64Constant, err
	}
	return obj.(int64), nil
}

func (i *injector) GetTaggedUint(tag string) (uint, error) {
	obj, err := i.getTaggedConstant(tag, uintConstantKind)
	if err != nil {
		return uintConstant, err
	}
	return obj.(uint), nil
}

func (i *injector) GetTaggedUint8(tag string) (uint8, error) {
	obj, err := i.getTaggedConstant(tag, uint8ConstantKind)
	if err != nil {
		return uint8Constant, err
	}
	return obj.(uint8), nil
}

func (i *injector) GetTaggedUint16(tag string) (uint16, error) {
	obj, err := i.getTaggedConstant(tag, uint16ConstantKind)
	if err != nil {
		return uint16Constant, err
	}
	return obj.(uint16), nil
}

func (i *injector) GetTaggedUint32(tag string) (uint32, error) {
	obj, err := i.getTaggedConstant(tag, uint32ConstantKind)
	if err != nil {
		return uint32Constant, err
	}
	return obj.(uint32), nil
}

func (i *injector) GetTaggedUint64(tag string) (uint64, error) {
	obj, err := i.getTaggedConstant(tag, uint64ConstantKind)
	if err != nil {
		return uint64Constant, err
	}
	return obj.(uint64), nil
}

func (i *injector) GetTaggedFloat32(tag string) (float32, error) {
	obj, err := i.getTaggedConstant(tag, float32ConstantKind)
	if err != nil {
		return float32Constant, err
	}
	return obj.(float32), nil
}

func (i *injector) GetTaggedFloat64(tag string) (float64, error) {
	obj, err := i.getTaggedConstant(tag, float64ConstantKind)
	if err != nil {
		return float64Constant, err
	}
	return obj.(float64), nil
}

func (i *injector) GetTaggedComplex64(tag string) (complex64, error) {
	obj, err := i.getTaggedConstant(tag, complex64ConstantKind)
	if err != nil {
		return complex64Constant, err
	}
	return obj.(complex64), nil
}

func (i *injector) GetTaggedComplex128(tag string) (complex128, error) {
	obj, err := i.getTaggedConstant(tag, complex128ConstantKind)
	if err != nil {
		return complex128Constant, err
	}
	return obj.(complex128), nil
}

func (i *injector) GetTaggedString(tag string) (string, error) {
	obj, err := i.getTaggedConstant(tag, stringConstantKind)
	if err != nil {
		return stringConstant, err
	}
	return obj.(string), nil
}

func (i *injector) getTaggedConstant(tag string, constantKind constantKind) (interface{}, error) {
	return i.get(newTaggedBindingKey(constantKind.reflectType(), tag))
}

func (i *injector) Call(function interface{}) ([]interface{}, error) {
	funcReflectType := reflect.TypeOf(function)
	if err := verifyIsFunc(funcReflectType); err != nil {
		return nil, err
	}
	bindingKeys := getParameterBindingKeysForFunc(funcReflectType)
	if err := i.validateBindingKeys(bindingKeys); err != nil {
		return nil, err
	}
	reflectValues, err := i.getReflectValues(bindingKeys)
	if err != nil {
		return nil, err
	}
	returnValues := reflect.ValueOf(function).Call(reflectValues)
	return reflectValuesToValues(returnValues), nil
}

func (i *injector) CallTagged(taggedFunction interface{}) ([]interface{}, error) {
	taggedFuncReflectType := reflect.TypeOf(taggedFunction)
	if err := verifyIsTaggedFunc(taggedFuncReflectType); err != nil {
		return nil, err
	}
	bindingKeys := getParameterBindingKeysForTaggedFunc(taggedFuncReflectType)
	if err := i.validateBindingKeys(bindingKeys); err != nil {
		return nil, err
	}
	reflectValues, err := i.getReflectValues(bindingKeys)
	if err != nil {
		return nil, err
	}
	structReflectValue := newStructReflectValue(taggedFuncReflectType.In(0))
	populateStructReflectValue(&structReflectValue, reflectValues)
	returnValues := reflect.ValueOf(taggedFunction).Call([]reflect.Value{structReflectValue})
	return reflectValuesToValues(returnValues), nil
}

func (i *injector) Populate(populateStructPtr interface{}) error {
	populateStructPtrReflectType := reflect.TypeOf(populateStructPtr)
	if err := verifyIsStructPtr(populateStructPtrReflectType); err != nil {
		return err
	}
	populateStructValue := reflect.Indirect(reflect.ValueOf(populateStructPtr))
	if err := verifyStructCanBePopulated(populateStructValue.Type()); err != nil {
		return err
	}
	bindingKeys := getStructFieldBindingKeys(populateStructValue.Type())
	if err := i.validateBindingKeys(bindingKeys); err != nil {
		return err
	}
	reflectValues, err := i.getReflectValues(bindingKeys)
	if err != nil {
		return err
	}
	populateStructReflectValue(&populateStructValue, reflectValues)
	return nil
}

func (i *injector) NewChildInjector(overridesType interface{}, modules ...Module) (Injector, error) {
	overrides, _ := i.Get(overridesType)
	if om, ok := overrides.(Module); ok {
		modules = []Module{Override(modules...).With(om)}
	}
	injector := &injector{i, make(map[bindingKey]resolvedBinding)}
	_, err := initInjector(injector, modules)
	if err != nil {
		return nil, err
	}
	return injector, nil
}

func (i *injector) get(bindingKey bindingKey) (interface{}, error) {
	binding, err := i.getBinding(bindingKey)
	if err != nil {
		return nil, err
	}
	return binding.get()
}

func (i *injector) getBinding(bindingKey bindingKey) (resolvedBinding, error) {
	// get binding from parent, if any, but not the injector itself
	if i.parent != nil && bindingKey.reflectType() != injectorReflectType {
		binding, err := i.parent.getBinding(bindingKey)
		if err == nil {
			return binding, nil
		}
	}
	// get local binding
	binding, ok := i.bindings[bindingKey]
	if !ok {
		return nil, errNoBinding.withTag("bindingKey", bindingKey)
	}
	return binding, nil
}

func (i *injector) getReflectValues(bindingKeys []bindingKey) ([]reflect.Value, error) {
	numBindingKeys := len(bindingKeys)
	reflectValues := make([]reflect.Value, numBindingKeys)
	for ii := 0; ii < numBindingKeys; ii++ {
		value, err := i.get(bindingKeys[ii])
		if err != nil {
			return nil, err
		}
		reflectValues[ii] = reflect.ValueOf(value)
	}
	return reflectValues, nil
}

func (i *injector) validateBindingKeys(bindingKeys []bindingKey) error {
	for _, bindingKey := range bindingKeys {
		if _, err := i.getBinding(bindingKey); err != nil {
			return err
		}
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
