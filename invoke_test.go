package dirt_test

import (
	"reflect"
	"testing"

	"github.com/h2570su/dirt"
)

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

func TestInvokeAs(t *testing.T) {
	type iFace interface{ Foo() }
	type ServiceA struct {
		dirt.Injectable
		iFace
	}
	type ServiceB struct{ dirt.Injectable }
	type ServiceC struct {
		dirt.Injectable
		iFace

		A *ServiceA `dirt:""`
	}

	scope := &dirt.Scope{}
	dirt.ProvideStruct[*ServiceA](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceB](dirt.Scoped(scope))
	dirt.ProvideStruct[*ServiceC](dirt.Scoped(scope))

	a, err := dirt.Invoke[*ServiceA](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}

	c, err := dirt.Invoke[*ServiceC](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}
	if c.A == nil {
		t.Fatal("dependency not injected")
	}
	if a != c.A {
		t.Fatal("different instance injected")
	}

	expectedAiFace, err := dirt.InvokeAs[iFace](dirt.Scoped(scope))
	if err != nil {
		t.Fatal(err)
	}
	if expectedAiFace != a {
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
