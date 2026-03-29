package dirt

type Options struct {
	Name  string
	Scope *Scope
}

type Option func(*Options)

func DefaultProvideOptions() *Options { return &Options{Scope: globalScope} }
func Named(name string) Option        { return func(opts *Options) { opts.Name = name } }
func Scoped(scope *Scope) Option      { return func(opts *Options) { opts.Scope = scope } }
