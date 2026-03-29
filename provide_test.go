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

func TestProvideStructNamed(t *testing.T) {
	type ServiceA struct {
		Injectable

		Config string
	}
	type ServiceB struct {
		Injectable

		AA *ServiceA `dirt:"name:aa"`
		AB *ServiceA `dirt:"name:ab"`
	}

	validate := func(t *testing.T, scope *Scope) {
		aa, err := Invoke[*ServiceA](Scoped(scope), Named("aa"))
		if err != nil {
			t.Fatal(err)
		}
		ab, err := Invoke[*ServiceA](Scoped(scope), Named("ab"))
		if err != nil {
			t.Fatal(err)
		}

		if aa == ab {
			t.Fatal("same instance injected for different names")
		}

		b, err := Invoke[*ServiceB](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
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
		scope := &Scope{}
		ProvideStruct[*ServiceA](Scoped(scope), Named("aa"))
		ProvideStruct[*ServiceA](Scoped(scope), Named("ab"))
		ProvideStruct[*ServiceB](Scoped(scope))

		validate(t, scope)
	})

	t.Run("B,A", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[*ServiceB](Scoped(scope))
		ProvideStruct[*ServiceA](Scoped(scope), Named("aa"))
		ProvideStruct[*ServiceA](Scoped(scope), Named("ab"))

		validate(t, scope)
	})
}
