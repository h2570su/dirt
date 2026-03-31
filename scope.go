package dirt

import (
	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/internal/scope/simple"
)

type typeNameKey = core.TypeNameKey

var globalScope = &Scope{}

type Scope = simple.Scope

func GlobalScope() *Scope { return globalScope }
