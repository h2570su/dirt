package dirt

type Options struct {
	Name  string
	Scope *Scope
}

type Option func(Options) Options

func defaultProvideOptions() Options { return Options{Scope: globalScope} }

func Named(name string) Option   { return func(o Options) Options { o.Name = name; return o } }
func Scoped(scope *Scope) Option { return func(o Options) Options { o.Scope = scope; return o } }
