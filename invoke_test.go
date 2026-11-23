package dirt_test

import (
	"reflect"
	"testing"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/internal/scope/simple"
)

func TestInvokeIndividual(t *testing.T) {
	type ServiceA struct {
		Config string
	}
	type ServiceB struct {
		A *ServiceA `dirt:""`
	}

	scope := simple.NewScope()
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
		Config string
	}
	type ServiceB struct {
		A *ServiceA `dirt:"individual"`
	}

	scope := simple.NewScope()
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

func TestInvokeAs(t *testing.T) {
	type iFace interface{ Foo() }
	type ServiceA struct {
		iFace
	}
	type ServiceB struct{}
	type ServiceC struct {
		iFace

		A *ServiceA `dirt:""`
	}

	scope := simple.NewScope()
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceC](dirt.Scoped(scope))

	a, err := dirt.Invoke[*ServiceA](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}

	c, err := dirt.Invoke[*ServiceC](dirt.Scoped(scope))
	switch {
	case err != nil:
		t.Fatal(err)
	case c.A == nil:
		t.Fatal("dependency not injected")
	case a != c.A:
		t.Fatal("different instance injected")
	}

	if expectedAiFace, err := dirt.InvokeAs[iFace](dirt.Scoped(scope)); err != nil {
		t.Fatal(err)
	} else if expectedAiFace != a {
		t.Fatal("got unexpected instance for iFace")
	}

	var gotIFaces []iFace
	for iFaceGot, err := range dirt.InvokeAsMany[iFace](dirt.Scoped(scope)) {
		if err != nil {
			t.Fatal(err)
		}
		gotIFaces = append(gotIFaces, iFaceGot)
	}
	if !reflect.DeepEqual([]iFace{a, c}, gotIFaces) {
		t.Fatal("got unexpected sequence of iFace")
	}
}

func TestInvokeAsWithProvideAs(t *testing.T) {
	type iFace interface{ Foo() }
	type ServiceA struct {
		iFace `dirt:"-"`
	}
	type ServiceB struct{}
	type ServiceC struct {
		iFace `dirt:"-"`

		A *ServiceA `dirt:""`
	}

	scope := simple.NewScope()
	dirt.ProvideAs[*ServiceA, iFace](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceC](dirt.Scoped(scope))

	a, err := dirt.Invoke[*ServiceA](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}

	c, err := dirt.Invoke[*ServiceC](dirt.Scoped(scope))
	switch {
	case err != nil:
		t.Fatal(err)
	case c.A == nil:
		t.Fatal("dependency not injected")
	case a != c.A:
		t.Fatal("different instance injected")
	}

	if expectedAiFace, err := dirt.InvokeAs[iFace](dirt.Scoped(scope)); err != nil {
		t.Fatal(err)
	} else if expectedAiFace != a {
		t.Fatal("got unexpected instance for iFace")
	}

	var gotIFaces []iFace
	for iFaceGot, err := range dirt.InvokeAsMany[iFace](dirt.Scoped(scope)) {
		if err != nil {
			t.Fatal(err)
		}
		gotIFaces = append(gotIFaces, iFaceGot)
	}
	if !reflect.DeepEqual([]iFace{a, c}, gotIFaces) {
		t.Fatal("got unexpected sequence of iFace")
	}
}
