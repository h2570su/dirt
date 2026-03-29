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
		ProvideStruct[*ServiceA](Scoped(scope))
		ProvideStruct[*ServiceB](Scoped(scope))

		b, err := Invoke[*ServiceB](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if b.A == nil {
			t.Fatal("dependency not injected")
		}
		a, err := Invoke[*ServiceA](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if a != b.A {
			t.Fatal("different instance injected")
		}
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[*ServiceB](Scoped(scope))
		ProvideStruct[*ServiceA](Scoped(scope))

		b, err := Invoke[*ServiceB](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if b.A == nil {
			t.Fatal("dependency not injected")
		}
		a, err := Invoke[*ServiceA](Scoped(scope))
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
	ProvideStruct[*ServiceA](Scoped(scope))
	ProvideStruct[*ServiceB](Scoped(scope))

	b.Run("Invoke", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_, err := Invoke[*ServiceB](Scoped(scope))
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
		var sbp **ServiceB
		for b.Loop() {
			a, _ := Invoke[*ServiceA](Scoped(scope))

			sb := &ServiceB{
				A: a,
			}
			_sbp := &sb
			sbp = _sbp
		}
		b.Context().Value(sbp)
	})
}
