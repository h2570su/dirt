package dirt

type Options struct {
	Name  string
	Scope *Scope
}

func DefaultProvideOptions() Options { return Options{Scope: globalScope} }
func Named(name string) Options      { return DefaultProvideOptions().Named(name) }
func Scoped(scope *Scope) Options    { return DefaultProvideOptions().Scoped(scope) }

func (o Options) amendDefault() Options {
	if o.Scope == nil {
		o.Scope = globalScope
	}
	return o
}

func (o Options) Named(name string) Options   { o.Name = name; return o }
func (o Options) Scoped(scope *Scope) Options { o.Scope = scope; return o }
