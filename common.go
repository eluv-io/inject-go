package inject

import (
	"reflect"
)

const (
	taggedFuncStructFieldTag = "inject"
)

// whitelisting types to make sure the framework works
func isSupportedBindingKeyReflectType(reflectType reflect.Type) bool {
	return isSupportedBindReflectType(reflectType) || isSupportedBindInterfaceReflectType(reflectType) || isSupportedBindConstantReflectType(reflectType)
}

func isSupportedNoTagParameterReflectType(reflectType reflect.Type) bool {
	return isSupportedBindReflectType(reflectType) || isSupportedBindInterfaceReflectType(reflectType)
}

func isSupportedBindReflectType(reflectType reflect.Type) bool {
	switch reflectType.Kind() {
	case reflect.Ptr:
		switch reflectType.Elem().Kind() {
		case reflect.Interface:
			return true
		case reflect.Struct:
			return true
		default:
			return false
		}
	case
		reflect.Struct,
		reflect.Slice,
		reflect.Map,
		reflect.Chan,
		reflect.Bool,
		reflect.String,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Func:
		return true
	default:
		return false
	}
}

func isSupportedBindInterfaceReflectType(reflectType reflect.Type) bool {
	switch reflectType.Kind() {
	case reflect.Ptr:
		switch reflectType.Elem().Kind() {
		case reflect.Interface:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func isSupportedBindConstantReflectType(reflectType reflect.Type) bool {
	_, ok := constantKindForReflectType(reflectType)
	return ok
}

func verifyIsFunc(funcReflectType reflect.Type) error {
	if !isFunc(funcReflectType) {
		return errNotFunction.withTag("funcReflectType", funcReflectType)
	}
	numIn := funcReflectType.NumIn()
	for i := 0; i < numIn; i++ {
		parameterReflectType := funcReflectType.In(i)
		if isInterface(parameterReflectType) {
			parameterReflectType = reflect.PtrTo(parameterReflectType)
		}
		if err := verifyParameterCanBeInjected(parameterReflectType, ""); err != nil {
			return err
		}
	}
	return nil
}

func verifyIsTaggedFunc(funcReflectType reflect.Type) error {
	if !isFunc(funcReflectType) {
		return errNotFunction.withTag("funcReflectType", funcReflectType)
	}
	if funcReflectType.NumIn() != 1 {
		return errTaggedParametersInvalid.withTag("funcReflectType", funcReflectType)
	}
	inReflectType := funcReflectType.In(0)
	if !isStruct(inReflectType) {
		return errTaggedParametersInvalid.withTag("funcReflectType", funcReflectType)
	}
	if inReflectType.Name() != "" {
		return errTaggedParametersInvalid.withTag("funcReflectType", funcReflectType)
	}
	return verifyStructCanBePopulated(inReflectType)
}

func verifyStructCanBePopulated(structReflectType reflect.Type) error {
	numFields := structReflectType.NumField()
	for i := 0; i < numFields; i++ {
		structFieldReflectType, tag := getStructFieldReflectTypeAndTag(structReflectType.Field(i))
		if err := verifyParameterCanBeInjected(structFieldReflectType, tag); err != nil {
			return err
		}
	}
	return nil
}

func verifyParameterCanBeInjected(parameterReflectType reflect.Type, tag string) error {
	if tag == "" && !isSupportedNoTagParameterReflectType(parameterReflectType) {
		return errNotSupportedYet.withTag("parameterReflectType", parameterReflectType)
	}
	if tag != "" && !isSupportedBindingKeyReflectType(parameterReflectType) {
		return errNotSupportedYet.withTag("parameterReflectType", parameterReflectType)
	}
	return nil
}

func getParameterBindingKeysForFunc(funcReflectType reflect.Type) []bindingKey {
	numIn := funcReflectType.NumIn()
	bindingKeys := make([]bindingKey, numIn)
	for i := 0; i < numIn; i++ {
		inReflectType := funcReflectType.In(i)
		if inReflectType.Kind() == reflect.Interface {
			inReflectType = reflect.PtrTo(inReflectType)
		}
		bindingKeys[i] = newBindingKey(inReflectType)
	}
	return bindingKeys
}

func getParameterBindingKeysForTaggedFunc(funcReflectType reflect.Type) []bindingKey {
	return getStructFieldBindingKeys(funcReflectType.In(0))
}

func getStructFieldBindingKeys(structReflectType reflect.Type) []bindingKey {
	numFields := structReflectType.NumField()
	bindingKeys := make([]bindingKey, numFields)
	for i := 0; i < numFields; i++ {
		structFieldReflectType, tag := getStructFieldReflectTypeAndTag(structReflectType.Field(i))
		if tag != "" {
			bindingKeys[i] = newTaggedBindingKey(structFieldReflectType, tag)
		} else {
			bindingKeys[i] = newBindingKey(structFieldReflectType)
		}
	}
	return bindingKeys
}

func getStructFieldReflectTypeAndTag(structField reflect.StructField) (reflect.Type, string) {
	structFieldReflectType := structField.Type
	if structFieldReflectType.Kind() == reflect.Interface {
		structFieldReflectType = reflect.PtrTo(structFieldReflectType)
	}
	return structFieldReflectType, structField.Tag.Get(taggedFuncStructFieldTag)
}

func getTaggedFuncStructReflectValue(structReflectType reflect.Type, reflectValues []reflect.Value) *reflect.Value {
	structReflectValue := reflect.Indirect(reflect.New(structReflectType))
	populateStructReflectValue(&structReflectValue, reflectValues)
	return &structReflectValue
}

func newStructReflectValue(structReflectType reflect.Type) reflect.Value {
	return reflect.Indirect(reflect.New(structReflectType))
}

func populateStructReflectValue(structReflectValue *reflect.Value, reflectValues []reflect.Value) {
	numReflectValues := len(reflectValues)
	for i := 0; i < numReflectValues; i++ {
		structReflectValue.Field(i).Set(reflectValues[i])
	}
}

func isInterfacePtr(reflectType reflect.Type) bool {
	return isPtr(reflectType) && isInterface(reflectType.Elem())
}

func isStructPtr(reflectType reflect.Type) bool {
	return isPtr(reflectType) && isStruct(reflectType.Elem())
}

func isInterface(reflectType reflect.Type) bool {
	return reflectType.Kind() == reflect.Interface
}

func isStruct(reflectType reflect.Type) bool {
	return reflectType.Kind() == reflect.Struct
}

func isPtr(reflectType reflect.Type) bool {
	return reflectType.Kind() == reflect.Ptr
}

func isFunc(reflectType reflect.Type) bool {
	return reflectType.Kind() == reflect.Func
}
