package dirt

import (
	"testing"
)

func TestProvideStruct(t *testing.T) {
	type ServiceA struct {
		Injectable

		Config string
	}
	type ServiceB struct {
		Injectable

		A *ServiceA `dirt:""`
	}

	t.Run("A,B", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[*ServiceA](scope)
		ProvideStruct[*ServiceB](scope)

		b, err := Invoke[*ServiceB](scope)
		if err != nil {
			t.Fatal(err)
		}
		if b.A == nil {
			t.Fatal("dependency not injected")
		}
		a, err := Invoke[*ServiceA](scope)
		if err != nil {
			t.Fatal(err)
		}
		if a != b.A {
			t.Fatal("different instance injected")
		}
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[*ServiceB](scope)
		ProvideStruct[*ServiceA](scope)

		b, err := Invoke[*ServiceB](scope)
		if err != nil {
			t.Fatal(err)
		}
		if b.A == nil {
			t.Fatal("dependency not injected")
		}
		a, err := Invoke[*ServiceA](scope)
		if err != nil {
			t.Fatal(err)
		}
		if a != b.A {
			t.Fatal("different instance injected")
		}
	})
}

func BenchmarkProvideStruct(b *testing.B) {
	type ServiceA struct {
		Injectable

		Config string
	}
	type ServiceB struct {
		Injectable

		A *ServiceA `dirt:""`
	}
	scope := &Scope{}
	ProvideStruct[*ServiceA](scope)
	ProvideStruct[*ServiceB](scope)

	b.Run("Invoke", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_, err := Invoke[*ServiceB](scope)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Ctor", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_, err := scope.registrations[1].ctor()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native Instantiation", func(b *testing.B) {
		b.ResetTimer()
		var sb *ServiceB
		for b.Loop() {
			sb = &ServiceB{
				A: &ServiceA{},
			}
		}
		b.Context().Value(sb)
	})
}
