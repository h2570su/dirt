package dirt

import "git.ttech.cc/astaroth/dirt/core"

type (
	Options = core.Options
	Option  = core.Option
)

var defaultOptions = Options{Scope: globalScope}

// Named specifies the name of the provided instance, which is used to distinguish different instances of the same type. By default, the name is empty string.
func Named(name string) Option { return core.Named(name) }

// Scoped specifies the scope to R/W the registration and instance. By default, the global scope is used.
func Scoped(scope core.IScope) Option { return core.Scoped(scope) }
