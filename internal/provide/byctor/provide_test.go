package byctor_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/internal/hook"
	"github.com/h2570su/dirt/internal/provide/byctor"
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

func BenchmarkInvoke(b *testing.B) {
	type ServiceA struct {
		Config string
	}
	type ServiceB struct {
		A *ServiceA
	}
	scope := &dirt.Scope{}
	byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(dirt.Scoped(scope)))
	byctor.ProvideCtor(func(a *ServiceA) *ServiceB { return &ServiceB{A: a} }, opt(dirt.Scoped(scope)))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()}
		for b.Loop() {
			_, err := scope.InvokeInstance(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		container := make(map[any]any)
		sa := &ServiceA{}
		sb := &ServiceB{A: sa}
		container[TypeKey[*ServiceA]{}] = sa
		container[TypeKey[*ServiceB]{}] = sb
		lock := &sync.RWMutex{}
		b.ResetTimer()
		for b.Loop() {
			lock.RLock()
			_, ok := container[TypeKey[*ServiceB]{}]
			lock.RUnlock()
			if !ok {
				b.Fatal("instance not found")
			}
		}
	})
}

func BenchmarkInstantiate(b *testing.B) {
	type ServiceA struct {
		Config string
	}
	type ServiceB struct {
		A *ServiceA
	}

	scope := &dirt.Scope{}
	byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(dirt.Scoped(scope)))
	byctor.ProvideCtor(func(a *ServiceA) *ServiceB { return &ServiceB{A: a} }, opt(dirt.Scoped(scope)))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()}
		for b.Loop() {
			_, err := scope.Instantiate(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native", func(b *testing.B) {
		b.ResetTimer()
		newA := func() *ServiceA { return &ServiceA{} }
		newB := func() *ServiceB { return &ServiceB{A: newA()} }

		b.ResetTimer()
		for b.Loop() {
			_ = newB()
		}
	})
}

func TestProvideCtor(t *testing.T) {
	type ServiceA struct {
		Config string
	}
	type ServiceAA struct {
		ConfigAnother string
	}
	type ServiceB struct {
		A  *ServiceA
		AA *ServiceAA
	}

	validate := func(t *testing.T, scope core.IScope) {
		a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceA]()})
		if err != nil {
			t.Fatal(err)
		}
		aa, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceAA]()})
		if err != nil {
			t.Fatal(err)
		}

		_b, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err != nil {
			t.Fatal(err)
		}
		b, _ := _b.(*ServiceB)
		if b.A == nil {
			t.Fatal("dependency not injected")
		}
		if a != b.A {
			t.Fatal("different instance injected")
		}
		if aa != b.AA {
			t.Fatal("different instance injected")
		}
	}

	t.Run("A,B", func(t *testing.T) {
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func() *ServiceAA { return &ServiceAA{} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func(a *ServiceA, aa *ServiceAA) *ServiceB { return &ServiceB{A: a, AA: aa} }, opt(core.Scoped(scope)))

		validate(t, scope)
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &simple.Scope{}
		byctor.ProvideCtor(func(a *ServiceA, aa *ServiceAA) *ServiceB { return &ServiceB{A: a, AA: aa} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func() *ServiceAA { return &ServiceAA{} }, opt(core.Scoped(scope)))

		validate(t, scope)
	})

	t.Run("B,A, b invoke first", func(t *testing.T) {
		scope := &simple.Scope{}
		byctor.ProvideCtor(func(a *ServiceA, aa *ServiceAA) *ServiceB { return &ServiceB{A: a, AA: aa} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func() *ServiceAA { return &ServiceAA{} }, opt(core.Scoped(scope)))

		_, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err != nil {
			t.Fatal(err)
		}

		validate(t, scope)
	})

	t.Run("nil error returned", func(t *testing.T) {
		type ServiceC struct {
			Config string
		}
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() (*ServiceC, error) { return &ServiceC{}, nil }, opt(core.Scoped(scope)))

		_c, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceC]()})
		if err != nil {
			t.Fatal(err)
		}
		c, ok := _c.(*ServiceC)
		if !ok {
			t.Fatal("unexpected type")
		}
		if c == nil {
			t.Fatal("unexpected nil instance")
		}
	})

	t.Run("some error returned", func(t *testing.T) {
		type ServiceD struct {
			Config string
		}
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() (*ServiceD, error) { return nil, errors.New("nope") }, opt(core.Scoped(scope)))

		_d, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceD]()})
		if err == nil || !strings.Contains(err.Error(), "nope") {
			t.Fatal("expected error but got nil or unexpected error message:", err)
		}
		if _d != nil {
			t.Fatal("expected nil instance but got non-nil")
		}
	})
}

func TestProvideCtorNamed(t *testing.T) {
	type ServiceA struct {
		Config string
	}

	validate := func(t *testing.T, scope core.IScope) {
		aa, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceA](), Name: "aa"})
		if err != nil {
			t.Fatal(err)
		}
		ab, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceA](), Name: "ab"})
		if err != nil {
			t.Fatal(err)
		}

		if aa == ab {
			t.Fatal("same instance injected for different names")
		}
	}

	t.Run("aa,ab", func(t *testing.T) {
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope), core.Named("aa")))
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope), core.Named("ab")))

		validate(t, scope)
	})
}

const HookTestMixinCalledValue = "hooked"

type HookTestMixin string

func (h *HookTestMixin) PostInject() error { *h = HookTestMixinCalledValue; return nil }

type HookTestErrorMixin string

func (h *HookTestErrorMixin) PostInject() error { return errors.New("hook error") }

func TestProvideStructWithHook(t *testing.T) {
	type ServiceA struct {
		HookTestMixin
	}

	t.Run("*T hook", func(t *testing.T) {
		var _ hook.IPostInjectHook = (*ServiceA)(nil) // Ensure *ServiceA implements IPostInjectHook
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope)))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceA]()})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(*ServiceA)
		if a.HookTestMixin != HookTestMixinCalledValue {
			t.Fatalf("hook not called, got: %s", a.HookTestMixin)
		}
	})
	t.Run("T hook", func(t *testing.T) {
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() *ServiceA { return &ServiceA{} }, opt(core.Scoped(scope)))

		_a, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceA]()})
		if err != nil {
			t.Fatal(err)
		}
		a, _ := _a.(*ServiceA)

		if a.HookTestMixin != HookTestMixinCalledValue {
			t.Fatalf("hook not called, got: %s", a.HookTestMixin)
		}
	})

	t.Run("hook error", func(t *testing.T) {
		type ServiceB struct {
			HookTestErrorMixin
		}
		scope := &simple.Scope{}
		byctor.ProvideCtor(func() *ServiceB { return &ServiceB{} }, opt(core.Scoped(scope)))

		_, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !strings.Contains(err.Error(), "PostInject hook error: hook error") {
			t.Fatalf("unexpected error message: %s", err.Error())
		}
	})
}

type (
	ServiceLoopABAa struct {
		Dep *ServiceLoopABAb
	}
	ServiceLoopABAb struct {
		Dep *ServiceLoopABAa
	}

	ServiceLoopABCAa struct {
		Dep *ServiceLoopABCAb
	}
	ServiceLoopABCAb struct {
		Dep *ServiceLoopABCAc
	}
	ServiceLoopABCAc struct {
		Dep *ServiceLoopABCAa
	}
)

func TestProvideStructLoop(t *testing.T) {
	validate := func() {
		if r := recover(); r != nil {
			if !strings.Contains(fmt.Sprint(r), "circular dependency detected") {
				t.Fatalf("unexpected panic message: %s", fmt.Sprint(r))
			}
		}
	}
	t.Run("loop A->B->A", func(t *testing.T) {
		defer validate()
		scope := &simple.Scope{}
		byctor.ProvideCtor(func(dep *ServiceLoopABAb) *ServiceLoopABAa { return &ServiceLoopABAa{Dep: dep} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func(dep *ServiceLoopABAa) *ServiceLoopABAb { return &ServiceLoopABAb{Dep: dep} }, opt(core.Scoped(scope)))
	})

	t.Run("loop A->B->C->A", func(t *testing.T) {
		defer validate()
		scope := &simple.Scope{}
		byctor.ProvideCtor(func(dep *ServiceLoopABCAb) *ServiceLoopABCAa { return &ServiceLoopABCAa{Dep: dep} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func(dep *ServiceLoopABCAc) *ServiceLoopABCAb { return &ServiceLoopABCAb{Dep: dep} }, opt(core.Scoped(scope)))
		byctor.ProvideCtor(func(dep *ServiceLoopABCAa) *ServiceLoopABCAc { return &ServiceLoopABCAc{Dep: dep} }, opt(core.Scoped(scope)))
	})
}
