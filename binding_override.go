package inject

// OverrideBuilder allows creating a new module with overridden bindings.
// See the Override() function for details
type OverrideBuilder interface {

	// With creates a new module with bindings from the source module and all
	// provided override modules. If a binding exists in both the source module
	// and an override module, the binding from the override module is retained.
	With(overrides ...Module) Module
}

type override struct {
	source *module
}

// With implements OverrideBuilder.With()
func (o *override) With(overrides ...Module) Module {
	m := newModule()
	addBindings(m, o.source)
	for _, om := range overrides {
		addBindings(m, om.(*module))
	}
	return m
}

func addBindings(target *module, source *module) {
	for k, v := range source.bindings {
		target.bindings[k] = v
	}
	// also add any binding errors from the source modules, because
	// error checking is only done at creation of the injector
	target.bindingErrors = append(target.bindingErrors, source.bindingErrors...)
	// plus the eager singletons
	target.eager = append(target.eager, source.eager...)
}

// Override returns a builder that allows replacing bindings of the given
// source module(s) with equivalent bindings (same binding keys) from other
// modules.
//
// This should only be used in tests in order to replace production bindings
// with test bindings:
//   module := Override(productionModule).With(testModule)
func Override(sources ...Module) OverrideBuilder {
	return &override{source: Combine(sources...).(*module)}
}

// Combine combines the given modules into a single module.
func Combine(modules ...Module) Module {
	if len(modules) == 0 {
		return nil
	}
	if len(modules) == 1 {
		return modules[0]
	}
	m := NewModule()
	m.Install(modules...)
	return m
}
