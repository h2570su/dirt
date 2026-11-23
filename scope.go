package dirt

import (
	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/hooks"
	"github.com/h2570su/dirt/internal/scope/simple"
)

type typeNameKey = core.TypeNameKey

var globalScope = func() *simple.Scope {
	s := simple.NewScope()
	s.HookOnCtorBuilt(hooks.HookCheckAppendPostInject)
	s.HookOnCtorBuilt(hooks.CheckPotentialInjectableMissingTag)
	return s
}()

type Scope = simple.Scope

func GlobalScope() *Scope { return globalScope }

func NewScopeWithGlobalRegistry() *Scope {
	return &Scope{IRegistry: globalScope, IContainer: &simple.Container{}}
}
