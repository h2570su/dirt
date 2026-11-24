package hooks_test

import (
	"reflect"
	"testing"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/hooks"
	"github.com/h2570su/dirt/internal/scope/simple"
)

func TestHookCheckAppendPostInject(t *testing.T) {
	type A string
	type B struct {
		AA *A `dirt:""`
	}
	hookExecuted := false
	scope := simple.NewScope()
	a := A("123")
	scope.HookOnInjecting(hooks.HookOnInjecting(func(from *A, dstT reflect.Type) *A {
		hookExecuted = true
		if from != any(&a) {
			t.Fatal("Expected interface not found")
		}
		return new(A)
	}))
	dirt.ProvideValue(&a, dirt.Scoped(scope))
	dirt.ProvideStruct[*B](dirt.Scoped(scope))
	if b, err := dirt.Invoke[*B](dirt.Scoped(scope)); err != nil {
		t.Fatalf("Failed to invoke [*B]: %v", err)
	} else if b.AA == &a {
		t.Fatal("Expected not equal")
	}
	if !hookExecuted {
		t.Fatal("Expected hook not executed")
	}
}
