package dirt

import (
	"testing"
)

func TestProvideStruct(t *testing.T) {
	type ServiceA struct {
		Injectable

		Config string
	}
	type ServiceAA struct {
		Injectable

		ConfigAnother string
	}
	type ServiceB struct {
		Injectable

		A  *ServiceA  `dirt:""`
		AA *ServiceAA `dirt:""`
	}

	validate := func(t *testing.T, scope *Scope) {
		a, err := Invoke[*ServiceA](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		aa, err := Invoke[*ServiceAA](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}

		b, err := Invoke[*ServiceB](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
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
		scope := &Scope{}
		ProvideStruct[*ServiceA](Scoped(scope))
		ProvideStruct[*ServiceAA](Scoped(scope))
		ProvideStruct[*ServiceB](Scoped(scope))

		validate(t, scope)
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[*ServiceB](Scoped(scope))
		ProvideStruct[*ServiceA](Scoped(scope))
		ProvideStruct[*ServiceAA](Scoped(scope))

		validate(t, scope)
	})
}

func TestProvideStructNested(t *testing.T) {
	type ServiceA struct {
		Injectable

		Config string
	}
	type ServiceAA struct {
		Injectable

		ConfigAnother string
	}
	type ServiceB struct {
		Injectable

		GroupA struct {
			Subclass

			A *ServiceA `dirt:""`
		}
		GroupB *struct {
			Subclass

			AA *ServiceAA `dirt:""`
		}
	}

	t.Run("A,B", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[*ServiceA](Scoped(scope))
		ProvideStruct[*ServiceAA](Scoped(scope))
		ProvideStruct[*ServiceB](Scoped(scope))

		b, err := Invoke[*ServiceB](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if b.GroupA.A == nil {
			t.Fatal("dependency not injected")
		}
		if b.GroupB.AA == nil {
			t.Fatal("dependency not injected")
		}
	})
}
