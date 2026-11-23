package byalias_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/internal/provide/byalias"
	"github.com/h2570su/dirt/internal/provide/byvalue"
	"github.com/h2570su/dirt/internal/scope/simple"
)

func opt(opts ...core.Option) core.Options {
	var opt core.Options
	for _, o := range opts {
		opt = o(opt)
	}
	return opt
}

type TypeKey[T any] struct{}

type (
	vType string
	pType string
	I     interface{ Foo() string }
)

func (a vType) Foo() string  { return string(a) }
func (b *pType) Foo() string { return string(*b) }

func BenchmarkInvoke(b *testing.B) {
	scope := simple.NewScope()
	byalias.ProvideAs[vType, I](opt(dirt.Scoped(scope)))
	byvalue.ProvideValue(vType("A"), opt(dirt.Scoped(scope)))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[vType]()}
		for b.Loop() {
			_, err := scope.InvokeInstance(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		container := make(map[any]any)
		va := vType("A")
		container[TypeKey[vType]{}] = va
		lock := &sync.RWMutex{}
		b.ResetTimer()
		for b.Loop() {
			lock.RLock()
			_, ok := container[TypeKey[vType]{}]
			lock.RUnlock()
			if !ok {
				b.Fatal("instance not found")
			}
		}
	})
}

func BenchmarkInstantiate(b *testing.B) {
	scope := simple.NewScope()
	byvalue.ProvideValue(vType("A"), opt(dirt.Scoped(scope)))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[vType]()}
		for b.Loop() {
			_, err := scope.Instantiate(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		b.ResetTimer()
		newA := func() vType { return vType("A") }

		b.ResetTimer()
		for b.Loop() {
			_ = newA()
		}
	})
}

func TestProvideAs(t *testing.T) {
	t.Run("value type", func(t *testing.T) {
		// Principal Check
		var _ I = vType("A") // Ensure vType implements I
		scope := simple.NewScope()
		byalias.ProvideAs[vType, I](opt(dirt.Scoped(scope)))
		byvalue.ProvideValue(vType("A"), opt(dirt.Scoped(scope)))

		_i, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[I]()})
		if err != nil {
			t.Fatal(err)
		} else if i, ok := _i.(I); !ok {
			t.Fatalf("unexpected type: %T", _i)
		} else if i.Foo() != "A" {
			t.Fatalf("unexpected value: %s", i.Foo())
		}
	})

	t.Run("pointer type", func(t *testing.T) {
		// Principal Check
		_ = func() I { b := "B"; return (*pType)(&b) } // Ensure *pType implements I
		scope := simple.NewScope()
		byalias.ProvideAs[*pType, I](opt(dirt.Scoped(scope)))
		byvalue.ProvideValue(func() *pType { b := "B"; return (*pType)(&b) }(), opt(dirt.Scoped(scope)))

		_i, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[I]()})
		if err != nil {
			t.Fatal(err)
		} else if i, ok := _i.(I); !ok {
			t.Fatalf("unexpected type: %T", _i)
		} else if i.Foo() != "B" {
			t.Fatalf("unexpected value: %s", i.Foo())
		}
	})
}

func TestProvideValueNamed(t *testing.T) {
	t.Run("named values", func(t *testing.T) {
		// Principal Check
		var _ I = vType("A") // Ensure vType implements I
		scope := simple.NewScope()
		byalias.ProvideAs[vType, I](opt(dirt.Scoped(scope), dirt.Named("a")))
		byalias.ProvideAs[vType, I](opt(dirt.Scoped(scope), dirt.Named("b")))
		byvalue.ProvideValue(vType("A"), opt(dirt.Scoped(scope), dirt.Named("a")))
		byvalue.ProvideValue(vType("B"), opt(dirt.Scoped(scope), dirt.Named("b")))

		if _i, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[I](), Name: "a"}); err != nil {
			t.Fatal(err)
		} else if i, ok := _i.(I); !ok {
			t.Fatalf("unexpected type: %T", _i)
		} else if i.Foo() != "A" {
			t.Fatalf("unexpected value: %s", i.Foo())
		}

		if _i, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[I](), Name: "b"}); err != nil {
			t.Fatal(err)
		} else if i, ok := _i.(I); !ok {
			t.Fatalf("unexpected type: %T", _i)
		} else if i.Foo() != "B" {
			t.Fatalf("unexpected value: %s", i.Foo())
		}
	})
	t.Run("named pointers", func(t *testing.T) {
		_ = func() I { b := "B"; return (*pType)(&b) } // Ensure *pType implements I
		scope := simple.NewScope()
		byalias.ProvideAs[*pType, I](opt(dirt.Scoped(scope), dirt.Named("a")))
		byalias.ProvideAs[*pType, I](opt(dirt.Scoped(scope), dirt.Named("b")))
		byvalue.ProvideValue(func() *pType { b := "A"; return (*pType)(&b) }(), opt(dirt.Scoped(scope), dirt.Named("a")))
		byvalue.ProvideValue(func() *pType { b := "B"; return (*pType)(&b) }(), opt(dirt.Scoped(scope), dirt.Named("b")))

		if _i, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[I](), Name: "a"}); err != nil {
			t.Fatal(err)
		} else if i, ok := _i.(I); !ok {
			t.Fatalf("unexpected type: %T", _i)
		} else if i.Foo() != "A" {
			t.Fatalf("unexpected value: %s", i.Foo())
		}

		if _i, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[I](), Name: "b"}); err != nil {
			t.Fatal(err)
		} else if i, ok := _i.(I); !ok {
			t.Fatalf("unexpected type: %T", _i)
		} else if i.Foo() != "B" {
			t.Fatalf("unexpected value: %s", i.Foo())
		}
	})
}
