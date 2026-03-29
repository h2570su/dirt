package dirt

import (
	"git.ttech.cc/astaroth/dirt/core"
	"git.ttech.cc/astaroth/dirt/internal/scope/simple"
)

type typeNameKey = core.TypeNameKey

var globalScope = &Scope{}

type Scope = simple.Scope
