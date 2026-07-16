package core

type Options struct {
	Name  string
	Scope IScope

	CheckInjectableMissingTagOnInvoke bool
}

type Option func(Options) Options

// Named specifies the name of the provided instance, which is used to distinguish different instances of the same type. By default, the name is empty string.
func Named(name string) Option { return func(o Options) Options { o.Name = name; return o } }

// Scoped specifies the scope to R/W the registration and instance. By default, the global scope is used.
func Scoped(scope IScope) Option { return func(o Options) Options { o.Scope = scope; return o } }

// CheckInjectableMissingTagOnInvoke specifies whether to check for missing tags on invoke.
func CheckInjectableMissingTagOnInvoke(check bool) Option {
	return func(o Options) Options { o.CheckInjectableMissingTagOnInvoke = check; return o }
}
