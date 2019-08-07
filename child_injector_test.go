package inject_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eluv-io/inject-go"
)

type GlobalScope string

type RestrictedScope string

type DependsOnGlobal struct {
	Global GlobalScope
}

type DependsOnRestricted struct {
	Restricted RestrictedScope
}

type DependsOnBoth struct {
	Global     GlobalScope
	Restricted RestrictedScope
}

const gs = GlobalScope("global")
const rs = RestrictedScope("restricted")

func TestChildInjector(t *testing.T) {
	gm := inject.NewModule()
	gm.Bind(GlobalScope("")).ToSingleton(gs)
	ginj, err := inject.NewInjector(gm)
	require.NoError(t, err)

	rm := inject.NewModule()
	rm.Bind(RestrictedScope("")).ToSingleton(rs)
	rinj, err := ginj.NewChildInjector(rm)
	require.NoError(t, err)

	var dog DependsOnGlobal
	var dor DependsOnRestricted
	var dob DependsOnBoth

	t.Run("populate global scope", func(t *testing.T) {
		err = ginj.Populate(&dog)
		require.NoError(t, err)
		require.Equal(t, gs, dog.Global)
	})

	t.Run("populate restricted scope", func(t *testing.T) {
		err = rinj.Populate(&dor)
		require.NoError(t, err)
		require.Equal(t, rs, dor.Restricted)
	})

	t.Run("populate both", func(t *testing.T) {
		err = rinj.Populate(&dob)
		require.NoError(t, err)
		require.Equal(t, gs, dob.Global)
		require.Equal(t, rs, dob.Restricted)
	})

	t.Run("get injector in global scope", func(t *testing.T) {
		var obj interface{}
		obj, err = ginj.Get((*inject.Injector)(nil))
		require.NoError(t, err)
		require.Equal(t, ginj, obj)
	})

	t.Run("get injector in restricted scope", func(t *testing.T) {
		var obj interface{}
		obj, err = rinj.Get((*inject.Injector)(nil))
		require.NoError(t, err)
		require.Equal(t, rinj, obj)
	})

	t.Run("error populating restricted dep in global scope", func(t *testing.T) {
		err = ginj.Populate(&dor)
		require.Error(t, err)
		require.Contains(t, err.Error(), "No binding for binding key")
	})

	t.Run("error populating global & restricted dep in global scope", func(t *testing.T) {
		err = ginj.Populate(&dob)
		require.Error(t, err)
		require.Contains(t, err.Error(), "No binding for binding key")
	})
}

func TestChildInjectorErrors(t *testing.T) {
	gm := inject.NewModule()
	gm.Bind(GlobalScope("")).ToSingleton(gs)
	ginj, err := inject.NewInjector(gm)
	require.NoError(t, err)

	t.Run("re-bind", func(t *testing.T) {
		rm := inject.NewModule()
		rm.Bind(GlobalScope("")).ToSingleton(GlobalScope("local"))
		rinj, err := ginj.NewChildInjector(rm)
		require.Error(t, err)
		require.Nil(t, rinj)
	})
}
