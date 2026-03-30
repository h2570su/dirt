package byvalue_test

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"git.ttech.cc/astaroth/dirt"
	"git.ttech.cc/astaroth/dirt/core"
	"git.ttech.cc/astaroth/dirt/internal/hook"
	"git.ttech.cc/astaroth/dirt/internal/provide/byvalue"
	"git.ttech.cc/astaroth/dirt/internal/scope/simple"
)

func opt(opts ...core.Option) core.Options {
	var opt core.Options
	for _, o := range opts {
		opt = o(opt)
	}
	return opt
}

type TypeKey[T any] struct{}

func BenchmarkInvoke(b *testing.B) {
	type ValueA int
	scope := &dirt.Scope{}
	byvalue.ProvideValue(ValueA(123), opt(dirt.Scoped(scope)))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[ValueA]()}
		for b.Loop() {
			_, err := scope.InvokeInstance(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		container := make(map[any]any)
		va := ValueA(123)
		container[TypeKey[ValueA]{}] = va
		lock := &sync.RWMutex{}
		b.ResetTimer()
		for b.Loop() {
			lock.RLock()
			_, ok := container[TypeKey[ValueA]{}]
			lock.RUnlock()
			if !ok {
				b.Fatal("instance not found")
			}
		}
	})
}

func BenchmarkInstantiate(b *testing.B) {
	type ValueA int

	scope := &dirt.Scope{}
	byvalue.ProvideValue(ValueA(123), opt(dirt.Scoped(scope)))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[ValueA]()}
		for b.Loop() {
			_, err := scope.Instantiate(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		b.ResetTimer()
		newA := func() ValueA { return ValueA(123) }

		b.ResetTimer()
		for b.Loop() {
			_ = newA()
		}
	})
}

func TestProvideValue(t *testing.T) {
	type ValueA int

	t.Run("value type", func(t *testing.T) {
		scope := &dirt.Scope{}
		byvalue.ProvideValue(ValueA(123), opt(dirt.Scoped(scope)))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[ValueA]()})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(ValueA)
		if a != 123 {
			t.Fatalf("unexpected value: %d", a)
		}
	})

	t.Run("pointer type", func(t *testing.T) {
		scope := &dirt.Scope{}
		ins := ValueA(123)
		byvalue.ProvideValue(&ins, opt(dirt.Scoped(scope)))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ValueA]()})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(*ValueA)
		if *a != 123 {
			t.Fatalf("unexpected value: %d", *a)
		}
		*a = 456
		if ins != 456 {
			t.Fatalf("value not updated: %d", ins)
		}
	})
}

func TestProvideValueNamed(t *testing.T) {
	type ValueA int

	t.Run("named values", func(t *testing.T) {
		scope := &dirt.Scope{}
		byvalue.ProvideValue(ValueA(123), opt(dirt.Scoped(scope), dirt.Named("a")))
		byvalue.ProvideValue(ValueA(456), opt(dirt.Scoped(scope), dirt.Named("b")))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[ValueA](), Name: "a"})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(ValueA)
		if a != 123 {
			t.Fatalf("unexpected value: %d", a)
		}

		_b, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[ValueA](), Name: "b"})
		if err != nil {
			t.Fatal(err)
		}
		b, _ := _b.(ValueA)
		if b != 456 {
			t.Fatalf("unexpected value: %d", b)
		}
	})
	t.Run("named pointers", func(t *testing.T) {
		scope := &dirt.Scope{}
		insA := ValueA(123)
		insB := ValueA(456)
		byvalue.ProvideValue(&insA, opt(dirt.Scoped(scope), dirt.Named("a")))
		byvalue.ProvideValue(&insB, opt(dirt.Scoped(scope), dirt.Named("b")))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ValueA](), Name: "a"})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(*ValueA)
		if *a != 123 {
			t.Fatalf("unexpected value: %d", *a)
		}
		if a != &insA {
			t.Fatal("different instance injected")
		}

		_b, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ValueA](), Name: "b"})
		if err != nil {
			t.Fatal(err)
		}
		b, _ := _b.(*ValueA)
		if *b != 456 {
			t.Fatalf("unexpected value: %d", *b)
		}
		if b != &insB {
			t.Fatal("different instance injected")
		}
	})
}

const HookTestStubValue = "hooked"

type HookTestStub string

func (h *HookTestStub) PostInject() error {
	if *h != "" {
		return errors.New(string(*h))
	}
	*h = HookTestStubValue
	return nil
}

func TestProvideStructWithHook(t *testing.T) {
	t.Run("*T hook", func(t *testing.T) {
		var _ hook.IPostInjectHook = (*HookTestStub)(nil) // Ensure *HookTestStub implements IPostInjectHook
		scope := &simple.Scope{}
		ins := HookTestStub("")
		byvalue.ProvideValue(&ins, opt(dirt.Scoped(scope)))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*HookTestStub]()})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(*HookTestStub)
		if *a != HookTestStubValue {
			t.Fatalf("hook not called, got: %s", *a)
		}
	})
	t.Run("T hook", func(t *testing.T) {
		scope := &simple.Scope{}
		ins := HookTestStub("")
		byvalue.ProvideValue(ins, opt(dirt.Scoped(scope)))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[HookTestStub]()})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(HookTestStub)
		if a != HookTestStubValue {
			t.Fatalf("hook not called, got: %s", a)
		}
	})

	t.Run("hook error", func(t *testing.T) {
		scope := &simple.Scope{}
		ins := HookTestStub("error")
		byvalue.ProvideValue(&ins, opt(dirt.Scoped(scope)))

		_, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*HookTestStub]()})
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
		if !strings.Contains(err.Error(), "error") {
			t.Fatalf("unexpected error message: %s", err.Error())
		}
	})
}
