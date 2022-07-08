package inject

import (
	"fmt"
	"reflect"
	"strings"
)

type module struct {
	bindings      map[bindingKey]binding
	bindingErrors []error
	eager         []*singletonBuilder
}

func newModule() *module {
	return &module{bindings: make(map[bindingKey]binding), bindingErrors: make([]error, 0)}
}

func (m *module) BindConstructor(fn interface{}) {
	m.bindAuto(false, fn)
}

func (m *module) BindSingletonConstructor(fn interface{}) SingletonBuilder {
	return m.bindAuto(true, fn)
}

func (m *module) bindAuto(singleton bool, fn interface{}) SingletonBuilder {
	t := reflect.TypeOf(fn)
	if err := verifyConstructorReflectType(nil, t); err != nil {
		m.addBindingError(err)
		return nil
	}
	out := t.Out(0)
	if out.Kind() == reflect.Interface {
		out = reflect.PtrTo(out)
	}
	if singleton {
		m.Bind(out).ToSingletonConstructor(fn)
		return newSingletonBuilder(m, out)
	}
	m.Bind(out).ToConstructor(fn)
	return nil
}

func (m *module) BindSingleton(singleton interface{}) {
	typ := reflect.TypeOf(singleton)
	if typ.Kind() == reflect.Interface {
		typ = reflect.PtrTo(typ)
	}
	m.Bind(typ).ToSingleton(singleton)
}

func (m *module) Bind(froms ...interface{}) Builder {
	if !m.verifySupportedTypes(froms, isSupportedBindReflectType) {
		return newNoOpBuilder()
	}
	return m.bind(newBindingKey, froms)
}

func (m *module) BindTagged(tag string, froms ...interface{}) Builder {
	if !m.verifyTag(tag) {
		return newNoOpBuilder()
	}
	if !m.verifySupportedTypes(froms, isSupportedBindReflectType) {
		return newNoOpBuilder()
	}
	return m.bind(func(fromReflectType reflect.Type) bindingKey { return newTaggedBindingKey(fromReflectType, tag) }, froms)
}

func (m *module) BindInterface(fromInterfaces ...interface{}) InterfaceBuilder {
	if !m.verifySupportedTypes(fromInterfaces, isSupportedBindInterfaceReflectType) {
		return newNoOpBuilder()
	}
	return m.bind(newBindingKey, fromInterfaces)
}

func (m *module) BindTaggedInterface(tag string, fromInterfaces ...interface{}) InterfaceBuilder {
	if !m.verifyTag(tag) {
		return newNoOpBuilder()
	}
	if !m.verifySupportedTypes(fromInterfaces, isSupportedBindInterfaceReflectType) {
		return newNoOpBuilder()
	}
	return m.bind(func(fromReflectType reflect.Type) bindingKey { return newTaggedBindingKey(fromReflectType, tag) }, fromInterfaces)
}

func (m *module) BindTaggedBool(tag string) Builder {
	return m.bindTaggedConstant(tag, boolConstantKind)
}

func (m *module) BindTaggedInt(tag string) Builder {
	return m.bindTaggedConstant(tag, intConstantKind)
}

func (m *module) BindTaggedInt8(tag string) Builder {
	return m.bindTaggedConstant(tag, int8ConstantKind)
}

func (m *module) BindTaggedInt16(tag string) Builder {
	return m.bindTaggedConstant(tag, int16ConstantKind)
}

func (m *module) BindTaggedInt32(tag string) Builder {
	return m.bindTaggedConstant(tag, int32ConstantKind)
}

func (m *module) BindTaggedInt64(tag string) Builder {
	return m.bindTaggedConstant(tag, int64ConstantKind)
}

func (m *module) BindTaggedUint(tag string) Builder {
	return m.bindTaggedConstant(tag, uintConstantKind)
}

func (m *module) BindTaggedUint8(tag string) Builder {
	return m.bindTaggedConstant(tag, uint8ConstantKind)
}

func (m *module) BindTaggedUint16(tag string) Builder {
	return m.bindTaggedConstant(tag, uint16ConstantKind)
}

func (m *module) BindTaggedUint32(tag string) Builder {
	return m.bindTaggedConstant(tag, uint32ConstantKind)
}

func (m *module) BindTaggedUint64(tag string) Builder {
	return m.bindTaggedConstant(tag, uint64ConstantKind)
}

func (m *module) BindTaggedFloat32(tag string) Builder {
	return m.bindTaggedConstant(tag, float32ConstantKind)
}

func (m *module) BindTaggedFloat64(tag string) Builder {
	return m.bindTaggedConstant(tag, float64ConstantKind)
}

func (m *module) BindTaggedComplex64(tag string) Builder {
	return m.bindTaggedConstant(tag, complex64ConstantKind)
}

func (m *module) BindTaggedComplex128(tag string) Builder {
	return m.bindTaggedConstant(tag, complex128ConstantKind)
}

func (m *module) BindTaggedString(tag string) Builder {
	return m.bindTaggedConstant(tag, stringConstantKind)
}

func (m *module) bindTaggedConstant(tag string, constantKind constantKind) Builder {
	if !m.verifyTag(tag) {
		return newNoOpBuilder()
	}
	if !m.verifySupportedTypes([]interface{}{constantKind.constant()}, isSupportedBindConstantReflectType) {
		return newNoOpBuilder()
	}
	return m.bind(func(fromReflectType reflect.Type) bindingKey { return newTaggedBindingKey(fromReflectType, tag) }, []interface{}{constantKind.constant()})
}

func (m *module) bind(newBindingKeyFunc func(reflect.Type) bindingKey, from []interface{}) InterfaceBuilder {
	lenFrom := len(from)
	if lenFrom == 0 {
		m.addBindingError(errNil)
		return newNoOpBuilder()
	}
	bindingKeys := make([]bindingKey, lenFrom)
	for i := 0; i < lenFrom; i++ {
		fromReflectType, ok := from[i].(reflect.Type)
		if !ok {
			fromReflectType = reflect.TypeOf(from[i])
		}
		if fromReflectType == nil {
			m.addBindingError(errNil)
			return newNoOpBuilder()
		}
		bindingKeys[i] = newBindingKeyFunc(fromReflectType)
	}
	return newBuilder(m, bindingKeys)
}

func (m *module) String() string {
	return fmt.Sprintf("module{%s}", strings.Join(m.keyValueStrings(), " "))
}

func (m *module) Install(others ...Module) {
	for _, mod := range others {
		m.install(mod.(*module))
	}
}
func (m *module) install(o *module) {
	m.bindingErrors = append(m.bindingErrors, o.bindingErrors...)
	m.eager = append(m.eager, o.eager...)
	for key, value := range o.bindings {
		m.setBinding(key, value)
	}
}

func (m *module) CallEagerly(function interface{}) {
	newSingletonBuilder(m, nil).EagerlyAndCall(function)
}

func (m *module) keyValueStrings() []string {
	strings := make([]string, len(m.bindings))
	i := 0
	for bindingKey, binding := range m.bindings {
		strings[i] = fmt.Sprintf("%s:%s", bindingKey.String(), binding.String())
		i++
	}
	return strings
}

func (m *module) addBindingError(err error) {
	m.bindingErrors = append(m.bindingErrors, err)
}

func (m *module) binding(bindingKey bindingKey) (binding, bool) {
	binding, ok := m.bindings[bindingKey]
	return binding, ok
}

func (m *module) setBinding(bindingKey bindingKey, binding binding) {
	foundBinding, ok := m.bindings[bindingKey]
	if ok {
		m.addBindingError(errAlreadyBound.withTag("bindingKey", bindingKey).withTag("foundBinding", foundBinding))
		return
	}
	m.bindings[bindingKey] = binding
}

func (m *module) verifyTag(tag string) bool {
	if tag == "" {
		m.addBindingError(errTagEmpty)
		return false
	}
	return true
}

func (m *module) verifySupportedTypes(froms []interface{}, isSupportedFunc func(reflect.Type) bool) bool {
	ok := true
	for _, from := range froms {
		// adds an error, so want to loop all the way through
		fromReflectType, ok := from.(reflect.Type)
		if !ok {
			fromReflectType = reflect.TypeOf(from)
		}
		if !m.verifySupportedType(fromReflectType, isSupportedFunc) {
			ok = false
		}
	}
	return ok
}

func (m *module) verifySupportedType(reflectType reflect.Type, isSupportedFunc func(reflect.Type) bool) bool {
	if !isSupportedFunc(reflectType) {
		m.addNotSupportedBindTypeError(reflectType)
		return false
	}
	return true
}

func (m *module) addNotSupportedBindTypeError(reflectType reflect.Type) {
	m.addBindingError(errNotSupportedBindType.withTag("reflectType", reflectType))
}
