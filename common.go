package inject

import (
	"reflect"
)

const (
	taggedFuncStructFieldTag = "injectTag"
)

// whitelisting types to make sure the framework works
func isSupportedBindingKeyReflectType(reflectType reflect.Type) bool {
	return isSupportedBindReflectType(reflectType) || isSupportedBindInterfaceReflectType(reflectType) || isSupportedBindConstantReflectType(reflectType)
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
	case reflect.Struct:
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
	switch reflectType.Kind() {
	case reflect.Bool:
		return true
	case reflect.Int:
		return true
	case reflect.Int8:
		return true
	case reflect.Int16:
		return true
	case reflect.Int32:
		return true
	case reflect.Int64:
		return true
	case reflect.Uint:
		return true
	case reflect.Uint8:
		return true
	case reflect.Uint16:
		return true
	case reflect.Uint32:
		return true
	case reflect.Uint64:
		return true
	case reflect.Float32:
		return true
	case reflect.Float64:
		return true
	case reflect.Complex64:
		return true
	case reflect.Complex128:
		return true
	case reflect.String:
		return true
	default:
		return false
	}
}

func verifyIsFunc(funcReflectType reflect.Type) error {
	if !isFunc(funcReflectType) {
		eb := newErrorBuilder(injectErrorTypeNotFunction)
		eb = eb.addTag("funcReflectType", funcReflectType)
		return eb.build()
	}
	numIn := funcReflectType.NumIn()
	for i := 0; i < numIn; i++ {
		err := verifyParameterCanBeInjected(funcReflectType.In(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyIsTaggedFunc(funcReflectType reflect.Type) error {
	if !isFunc(funcReflectType) {
		eb := newErrorBuilder(injectErrorTypeNotFunction)
		eb = eb.addTag("funcReflectType", funcReflectType)
		return eb.build()
	}
	if funcReflectType.NumIn() != 1 {
		eb := newErrorBuilder(injectErrorTypeTaggedParametersInvalid)
		eb = eb.addTag("funcReflectType", funcReflectType)
		return eb.build()
	}
	inReflectType := funcReflectType.In(0)
	if !isStruct(inReflectType) {
		eb := newErrorBuilder(injectErrorTypeTaggedParametersInvalid)
		eb = eb.addTag("funcReflectType", funcReflectType)
		return eb.build()
	}
	if inReflectType.Name() != "" {
		eb := newErrorBuilder(injectErrorTypeTaggedParametersInvalid)
		eb = eb.addTag("funcReflectType", funcReflectType)
		return eb.build()
	}
	numFields := inReflectType.NumField()
	for i := 0; i < numFields; i++ {
		err := verifyParameterCanBeInjected(inReflectType.Field(i).Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyParameterCanBeInjected(parameterReflectType reflect.Type) error {
	if isInterface(parameterReflectType) {
		parameterReflectType = reflect.PtrTo(parameterReflectType)
	}
	if !isSupportedBindingKeyReflectType(parameterReflectType) {
		eb := newErrorBuilder(injectErrorTypeNotSupportedYet)
		eb.addTag("parameterReflectType", parameterReflectType)
		return eb.build()
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
		structField := structReflectType.Field(i)
		structFieldReflectType := structField.Type
		if structFieldReflectType.Kind() == reflect.Interface {
			structFieldReflectType = reflect.PtrTo(structFieldReflectType)
		}
		tag := structField.Tag.Get(taggedFuncStructFieldTag)
		if tag != "" {
			bindingKeys[i] = newTaggedBindingKey(structFieldReflectType, tag)
		} else {
			bindingKeys[i] = newBindingKey(structFieldReflectType)
		}
	}
	return bindingKeys
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