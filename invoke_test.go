package dirt_test

import (
	"testing"

	"git.ttech.cc/astaroth/dirt"
)

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
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		opts := []dirt.Option{dirt.Scoped(scope)}
		for b.Loop() {
			_, err := dirt.Invoke[*ServiceB](opts...)
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
		b.ResetTimer()
		for b.Loop() {
			_, ok := container[TypeKey[*ServiceB]{}]
			if !ok {
				b.Fatal("instance not found")
			}
		}
	})
}

func BenchmarkInvokeIndividual(b *testing.B) {
	type ServiceA struct {
		dirt.Injectable

		Config string
	}
	type ServiceB struct {
		dirt.Injectable

		A *ServiceA `dirt:"individual"`
	}
	scope := &dirt.Scope{}
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))

	b.Run("dirt", func(b *testing.B) {
		b.ResetTimer()
		opts := []dirt.Option{dirt.Scoped(scope)}
		for b.Loop() {
			_, err := dirt.InvokeIndividual[*ServiceB](opts...)
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

func TestInvokeIndividual(t *testing.T) {
	type ServiceA struct {
		dirt.Injectable

		Config string
	}
	type ServiceB struct {
		dirt.Injectable

		A *ServiceA `dirt:""`
	}

	scope := &dirt.Scope{}
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))

	b1, err := dirt.InvokeIndividual[*ServiceB](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}
	b2, err := dirt.InvokeIndividual[*ServiceB](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}

	if b1 == b2 {
		t.Fatal("InvokeIndividual should return different instances")
	}
}

func TestInvokeIndividualInner(t *testing.T) {
	type ServiceA struct {
		dirt.Injectable

		Config string
	}
	type ServiceB struct {
		dirt.Injectable

		A *ServiceA `dirt:"individual"`
	}

	scope := &dirt.Scope{}
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))

	b1, err := dirt.InvokeIndividual[*ServiceB](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}
	b2, err := dirt.InvokeIndividual[*ServiceB](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}

	if b1 == b2 {
		t.Fatal("InvokeIndividual should return different instances")
	}
	if b1.A == b2.A {
		t.Fatal("InvokeIndividual should return different instances for dependencies")
	}
}
