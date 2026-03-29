package dirt

import (
	"errors"
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

const HookTestMixinCalledValue = "hooked"

type HookTestMixin string

func (h *HookTestMixin) PostInject() error { *h = HookTestMixinCalledValue; return nil }

type HookTestErrorMixin string

func (h *HookTestErrorMixin) PostInject() error { return errors.New("hook error") }

func TestProvideStructWithHook(t *testing.T) {
	type ServiceA struct {
		Injectable

		HookTestMixin
	}

	t.Run("*T hook", func(t *testing.T) {
		var _ IPostInjectHook = (*ServiceA)(nil) // Ensure *ServiceA implements IPostInjectHook
		scope := &Scope{}
		ProvideStruct[*ServiceA](Scoped(scope))

		a, err := Invoke[*ServiceA](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if a.HookTestMixin != HookTestMixinCalledValue {
			t.Fatalf("hook not called, got: %s", a.HookTestMixin)
		}
	})
	t.Run("T hook", func(t *testing.T) {
		scope := &Scope{}
		ProvideStruct[ServiceA](Scoped(scope))

		a, err := Invoke[ServiceA](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if a.HookTestMixin != HookTestMixinCalledValue {
			t.Fatalf("hook not called, got: %s", a.HookTestMixin)
		}
	})

	t.Run("hook error", func(t *testing.T) {
		type ServiceB struct {
			Injectable

			HookTestErrorMixin
		}
		scope := &Scope{}
		ProvideStruct[*ServiceB](Scoped(scope))

		_, err := Invoke[*ServiceB](Scoped(scope))
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err.Error() != "PostInject hook error: hook error" {
			t.Fatalf("unexpected error message: %s", err.Error())
		}
	})
}

func TestProvideStructOptional(t *testing.T) {
	type ServiceSuccess struct {
		Injectable

		HookTestMixin
	}

	type ServiceFail struct {
		Injectable

		HookTestErrorMixin
	}

	t.Run("optional", func(t *testing.T) {
		type Service struct {
			Injectable

			Master *ServiceSuccess `dirt:""`
			Slave  *ServiceFail    `dirt:"optional"`
		}
		scope := &Scope{}
		ProvideStruct[*ServiceSuccess](Scoped(scope))
		ProvideStruct[*ServiceFail](Scoped(scope))
		ProvideStruct[*Service](Scoped(scope))

		s, err := Invoke[*Service](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
	t.Run("optional,reversed", func(t *testing.T) {
		type Service struct {
			Injectable

			Slave  *ServiceFail    `dirt:"optional"`
			Master *ServiceSuccess `dirt:""`
		}
		scope := &Scope{}
		ProvideStruct[*ServiceSuccess](Scoped(scope))
		ProvideStruct[*ServiceFail](Scoped(scope))
		ProvideStruct[*Service](Scoped(scope))

		s, err := Invoke[*Service](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
	t.Run("optional,individual", func(t *testing.T) {
		type Service struct {
			Injectable

			Master *ServiceSuccess `dirt:"individual"`
			Slave  *ServiceFail    `dirt:"optional,individual"`
		}
		scope := &Scope{}
		ProvideStruct[*ServiceSuccess](Scoped(scope))
		ProvideStruct[*ServiceFail](Scoped(scope))
		ProvideStruct[*Service](Scoped(scope))

		s, err := Invoke[*Service](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
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
			Injectable

			Master *ServiceSuccess  `dirt:""`
			Slave  *ServiceNotExist `dirt:"optional"`
		}
		scope := &Scope{}
		ProvideStruct[*ServiceSuccess](Scoped(scope))
		ProvideStruct[*Service](Scoped(scope))

		s, err := Invoke[*Service](Scoped(scope))
		if err != nil {
			t.Fatal(err)
		}
		if s.Master == nil {
			t.Fatal("Master should be injected")
		}
		if s.Slave != nil {
			t.Fatal("Slave should not be injected due to hook error, but should not cause overall failure")
		}
	})
}
