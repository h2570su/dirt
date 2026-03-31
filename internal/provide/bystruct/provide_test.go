package bystruct_test

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
	"github.com/h2570su/dirt/internal/provide/bystruct"
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
		dirt.Injectable

		Config string
	}
	type ServiceB struct {
		dirt.Injectable

		A *ServiceA `dirt:""`
	}
	scope := &dirt.Scope{}
	bystruct.ProvideStruct[*ServiceA](opt(dirt.Scoped(scope)))
	bystruct.ProvideStruct[*ServiceB](opt(dirt.Scoped(scope)))

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
		dirt.Injectable

		Config string
	}
	type ServiceB struct {
		dirt.Injectable

		A *ServiceA `dirt:"individual"`
	}
	type ServiceBa struct {
		dirt.Injectable

		a *ServiceA `dirt:"individual"` //nolint:unused
	}

	scope := &dirt.Scope{}
	bystruct.ProvideStruct[*ServiceA](opt(dirt.Scoped(scope)))
	bystruct.ProvideStruct[*ServiceB](opt(dirt.Scoped(scope)))
	bystruct.ProvideStruct[*ServiceBa](opt(dirt.Scoped(scope)))

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
	b.Run("dirt,unexported", func(b *testing.B) {
		b.ResetTimer()
		key := core.TypeNameKey{Type: reflect.TypeFor[*ServiceBa]()}
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

func TestProvideStruct(t *testing.T) {
	type ServiceA struct {
		bystruct.Injectable

		Config string
	}
	type ServiceAA struct {
		bystruct.Injectable

		ConfigAnother string
	}
	type ServiceB struct {
		bystruct.Injectable

		A  *ServiceA  `dirt:""`
		AA *ServiceAA `dirt:""`
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
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceAA](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))

		validate(t, scope)
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceAA](opt(core.Scoped(scope)))

		validate(t, scope)
	})
}

func TestProvideStructUnexported(t *testing.T) {
	type ServiceA struct {
		bystruct.Injectable

		Config string
	}
	type ServiceB struct {
		bystruct.Injectable

		a *ServiceA `dirt:""`
	}

	t.Run("A,B", func(t *testing.T) {
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope)))

		_b, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err != nil {
			t.Fatal(err)
		}
		b, _ := _b.(*ServiceB)
		if b.a == nil {
			t.Fatal("dependency not injected")
		}
	})
}

func TestProvideStructNested(t *testing.T) {
	type ServiceA struct {
		bystruct.Injectable

		Config string
	}
	type ServiceAA struct {
		bystruct.Injectable

		ConfigAnother string
	}
	type ServiceB struct {
		bystruct.Injectable

		GroupA struct {
			bystruct.Subclass

			A *ServiceA `dirt:""`
		}
		GroupB *struct {
			bystruct.Subclass

			AA *ServiceAA `dirt:""`
		}
	}

	t.Run("A,B", func(t *testing.T) {
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceAA](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))

		_b, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err != nil {
			t.Fatal(err)
		}
		b, _ := _b.(*ServiceB)
		if b.GroupA.A == nil {
			t.Fatal("dependency not injected")
		}
		if b.GroupB.AA == nil {
			t.Fatal("dependency not injected")
		}
	})
}

func TestProvideStructNamed(t *testing.T) {
	type ServiceA struct {
		bystruct.Injectable

		Config string
	}
	type ServiceB struct {
		bystruct.Injectable

		AA *ServiceA `dirt:"name:aa"`
		AB *ServiceA `dirt:"name:ab"`
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

		_b, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err != nil {
			t.Fatal(err)
		}
		b, _ := _b.(*ServiceB)
		if b.AA == nil || b.AB == nil {
			t.Fatal("dependency not injected")
		}
		if aa != b.AA {
			t.Fatal("different instance injected")
		}
		if ab != b.AB {
			t.Fatal("different instance injected")
		}
		if aa == b.AB || ab == b.AA {
			t.Fatal("same instance injected for different names")
		}
	}

	t.Run("A,B", func(t *testing.T) {
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope), core.Named("aa")))
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope), core.Named("ab")))
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))

		validate(t, scope)
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope), core.Named("aa")))
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope), core.Named("ab")))

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
		bystruct.Injectable

		HookTestMixin
	}

	t.Run("*T hook", func(t *testing.T) {
		var _ hook.IPostInjectHook = (*ServiceA)(nil) // Ensure *ServiceA implements IPostInjectHook
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope)))

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
		bystruct.ProvideStruct[*ServiceA](opt(core.Scoped(scope)))

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
			bystruct.Injectable

			HookTestErrorMixin
		}
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceB](opt(core.Scoped(scope)))

		_, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*ServiceB]()})
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !strings.Contains(err.Error(), "PostInject hook error: hook error") {
			t.Fatalf("unexpected error message: %s", err.Error())
		}
	})
}

func TestProvideStructOptional(t *testing.T) {
	type ServiceSuccess struct {
		bystruct.Injectable

		HookTestMixin
	}

	type ServiceFail struct {
		bystruct.Injectable

		HookTestErrorMixin
	}

	t.Run("optional", func(t *testing.T) {
		type Service struct {
			bystruct.Injectable

			Master *ServiceSuccess `dirt:""`
			Slave  *ServiceFail    `dirt:"optional"`
		}
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceSuccess](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceFail](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*Service](opt(core.Scoped(scope)))

		_s, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*Service]()})
		if err != nil {
			t.Fatal(err)
		}
		s, _ := _s.(*Service)
		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
	t.Run("optional,reversed", func(t *testing.T) {
		type Service struct {
			bystruct.Injectable

			Slave  *ServiceFail    `dirt:"optional"`
			Master *ServiceSuccess `dirt:""`
		}
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceSuccess](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceFail](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*Service](opt(core.Scoped(scope)))

		_s, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*Service]()})
		if err != nil {
			t.Fatal(err)
		}
		s, _ := _s.(*Service)

		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
	t.Run("optional,individual", func(t *testing.T) {
		type Service struct {
			bystruct.Injectable

			Master *ServiceSuccess `dirt:"individual"`
			Slave  *ServiceFail    `dirt:"optional,individual"`
		}
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceSuccess](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceFail](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*Service](opt(core.Scoped(scope)))

		_s, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*Service]()})
		if err != nil {
			t.Fatal(err)
		}
		s, _ := _s.(*Service)

		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
	t.Run("optional,one-not-exist", func(t *testing.T) {
		type ServiceNotExist struct{}
		type Service struct {
			bystruct.Injectable

			Master *ServiceSuccess  `dirt:""`
			Slave  *ServiceNotExist `dirt:"optional"`
		}
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceSuccess](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*Service](opt(core.Scoped(scope)))

		_s, err := scope.InvokeInstance(core.TypeNameKey{Type: reflect.TypeFor[*Service]()})
		if err != nil {
			t.Fatal(err)
		}
		s, _ := _s.(*Service)

		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
}

type (
	ServiceLoopABAa struct {
		bystruct.Injectable

		Dep *ServiceLoopABAb `dirt:""`
	}
	ServiceLoopABAb struct {
		bystruct.Injectable

		Dep *ServiceLoopABAa `dirt:""`
	}

	ServiceLoopABCAa struct {
		bystruct.Injectable

		Dep *ServiceLoopABCAb `dirt:""`
	}
	ServiceLoopABCAb struct {
		bystruct.Injectable

		Dep *ServiceLoopABCAc `dirt:""`
	}
	ServiceLoopABCAc struct {
		bystruct.Injectable

		Dep *ServiceLoopABCAa `dirt:""`
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
		bystruct.ProvideStruct[*ServiceLoopABAa](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceLoopABAb](opt(core.Scoped(scope)))
	})
	t.Run("loop A->B,individual", func(t *testing.T) {
		defer validate()
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceLoopABAa](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceLoopABAa](opt(core.Scoped(scope)))
	})
	t.Run("loop A->B->C->A", func(t *testing.T) {
		defer validate()
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceLoopABCAa](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceLoopABCAb](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceLoopABCAc](opt(core.Scoped(scope)))
	})
	t.Run("loop A->B->C->A,individual", func(t *testing.T) {
		defer validate()
		scope := &simple.Scope{}
		bystruct.ProvideStruct[*ServiceLoopABCAa](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceLoopABCAb](opt(core.Scoped(scope)))
		bystruct.ProvideStruct[*ServiceLoopABCAa](opt(core.Scoped(scope)))
	})
}
