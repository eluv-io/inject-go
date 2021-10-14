package inject

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	injectErrorTypeNil                            = "Parameter is nil"
	injectErrorTypeReflectTypeNil                 = "reflect.TypeOf() returns nil"
	injectErrorTypeNotSupportedYet                = "Binding type not supported yet, feel free to help!"
	injectErrorTypeNotAssignable                  = "Binding not assignable"
	injectErrorTypeConstructorReturnValuesInvalid = "Constructor can only have two return values, the first providing the value, the second being an error"
	injectErrorTypeIntermediateBinding            = "Trying to get for an intermediate binding"
	injectErrorTypeFinalBinding                   = "Trying to get bindingKey for a final binding"
	injectErrorTypeCannotCastModule               = "Cannot cast Module to internal module type"
	injectErrorTypeNoBinding                      = "No binding for binding key"
	injectErrorTypeNoFinalBinding                 = "No final binding for binding key"
	injectErrorTypeAlreadyBound                   = "Already found a binding for this binding key"
	injectErrorTypeTagEmpty                       = "Tag empty"
	injectErrorTypeTaggedParametersInvalid        = "Tagged function must have one anonymous struct parameter"
	injectErrorTypeNotFunction                    = "Argument is not a function"
	injectErrorTypeNotInterfacePtr                = "Value is not an interface pointer"
	injectErrorTypeNotStructPtr                   = "Value is not a struct pointer"
	injectErrorTypeNotSupportedBindType           = "Type is not supported for this binding method"
	injectErrorTypeBindingErrors                  = "Errors with bindings"
	injectErrorTypeWrapped                        = "Wrapped standard error"
	injectErrorTypeConstructorCall                = "Constructor call failed"
)

var (
	errNil                            = newInjectError(injectErrorTypeNil)
	errReflectTypeNil                 = newInjectError(injectErrorTypeReflectTypeNil)
	errNotSupportedYet                = newInjectError(injectErrorTypeNotSupportedYet)
	errNotAssignable                  = newInjectError(injectErrorTypeNotAssignable)
	errConstructorReturnValuesInvalid = newInjectError(injectErrorTypeConstructorReturnValuesInvalid)
	errIntermediateBinding            = newInjectError(injectErrorTypeIntermediateBinding)
	errFinalBinding                   = newInjectError(injectErrorTypeFinalBinding)
	errCannotCastModule               = newInjectError(injectErrorTypeCannotCastModule)
	errNoBinding                      = newInjectError(injectErrorTypeNoBinding)
	errNoFinalBinding                 = newInjectError(injectErrorTypeNoFinalBinding)
	errAlreadyBound                   = newInjectError(injectErrorTypeAlreadyBound)
	errTagEmpty                       = newInjectError(injectErrorTypeTagEmpty)
	errTaggedParametersInvalid        = newInjectError(injectErrorTypeTaggedParametersInvalid)
	errNotFunction                    = newInjectError(injectErrorTypeNotFunction)
	errNotInterfacePtr                = newInjectError(injectErrorTypeNotInterfacePtr)
	errNotStructPtr                   = newInjectError(injectErrorTypeNotStructPtr)
	errNotSupportedBindType           = newInjectError(injectErrorTypeNotSupportedBindType)
	errBindingErrors                  = newInjectError(injectErrorTypeBindingErrors)
	errBindingWrapped                 = newInjectError(injectErrorTypeWrapped)
	errConstructorCall                = newInjectError(injectErrorTypeConstructorCall)
)

type injectError struct {
	errorType string
	tags      injectErrorTags
}

func newInjectError(errorType string) *injectError {
	return &injectError{errorType, make([]*injectErrorTag, 0)}
}

func (i *injectError) Error() string {
	value := fmt.Sprintf("inject: %s", i.errorType)
	if len(i.tags) == 0 {
		return value
	}
	return fmt.Sprintf("%s\n\t%s", value, strings.ReplaceAll(i.tags.String(), "\n", "\n\t"))
}

func (i *injectError) withTag(key string, value interface{}, nostack ...bool) *injectError {
	stack := true
	if len(nostack) > 0 && nostack[0] {
		stack = false
	}
	return &injectError{i.errorType, append(i.tags, newInjectErrorTag(key, value, stack))}
}

type injectErrorTag struct {
	key     string
	value   interface{}
	callers []uintptr
}

func newInjectErrorTag(key string, value interface{}, stack bool) *injectErrorTag {
	var callStack []uintptr
	if stack {
		callStack = callers(3)
	}
	return &injectErrorTag{key, value, callStack}
}

func (t *injectErrorTag) String() string {
	stack := ""
	if len(t.callers) > 0 {
		stack = "\n" + printStack(t.callers)
	}
	if stringer, ok := t.value.(fmt.Stringer); ok {
		return fmt.Sprintf("%s:%s%s", t.key, stringer.String(), stack)
	}
	return fmt.Sprintf("%s:%s%s", t.key, t.value, stack)
}

func printStack(callers []uintptr) string {
	sb := &strings.Builder{}
	frames := runtime.CallersFrames(callers)
	show := false
	for {
		frame, more := frames.Next()
		name := frame.Function
		if name == "" {
			show = true
			fmt.Fprintf(sb, "\t%#x\n", frame.PC)
		} else if name != "runtime.goexit" && (show || !strings.HasPrefix(name, "runtime.")) {
			// Hide runtime.goexit and any runtime functions at the beginning.
			// This is useful mainly for allocation traces.
			show = true
			fmt.Fprintf(sb, "\t%s:%d\t%s\n", frame.File, frame.Line, name)
		}
		if !more {
			break
		}
	}
	return sb.String()
}

type injectErrorTags []*injectErrorTag

func (ts injectErrorTags) String() string {
	if len(ts) == 0 {
		return ""
	}
	s := make([]string, len(ts))
	for i, tag := range ts {
		s[i] = tag.String()
	}
	return fmt.Sprintf("%s", strings.Join(s, "\n"))
}

func unwrap(err error) *injectError {
	if e, ok := err.(*injectError); ok {
		return e
	}
	return errBindingWrapped.withTag("err", err)
}

func callers(skip int) []uintptr {
	var pcs [512]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	return pcs[:n]
}
