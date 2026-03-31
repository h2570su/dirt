# dirt is a DI framework for Go 1.24+, inspired by samber/do/v2

DIRT stands for "Dependency Injection powered by Reflection and generic Types".

To install:

```bash
go get github.com/h2570su/dirt
```

## Why

I like the design of samber/do/v2, but in my experience with it, I found some limitations that make it less convenient to use in certain scenarios:

1. Hook doesn't give access to the instance.
1.1. No way to do post-inject initialization, which is common for struct with non-injectable fields after InjectStruct(), which I really like to use InjectStruct for field injection.
1.1.1. Even the Provide(NewXXX(i do.Injector)) is enough for custom injection logic, but it's annoying to write a lot of iferr(long) or MustInvoke(panic) in the ctor.
2. All Provide() must be hand-written(because of golang generics limitation), cannot be provided on run-time, which is useful for some dynamic initialization, e.g. based on config file or env vars.

## Thoughts

The inspiration of golang at struct is zero-value initialization, which doesn't require any ctor, even New()... pattern is common.
So I want to use injection by field as the major way, and ctor is just a supplement for some special cases, e.g. when the struct is not injectable, or when the struct is from external package and cannot be modified.

## Features

### Registration API

- `ProvideStruct[T any](opts ...dirt.Option)`
  - Register an injectable struct type by field analysis.
  - Return a dummy value so you can do `var _ = dirt.ProvideStruct[MyType]()`.
  - If T is not struct or struct pointer, it will **panic** immediately.
  - Field Tag (`dirt`). Only tagged fields are injected by default.
    - `dirt:"-"`: ignore this field.
    - `dirt:"name:xxx"`: inject dependency by name.
    - `dirt:"optional"`: dependency is optional.
    - `dirt:"individual"`: resolve this field as a fresh instance.
    - Combine options with `,`, e.g. `dirt:"name:xxx,optional"`.
    - **Interface dependency fields are not supported currently.**
- `ProvideCtor(fn any, opts ...dirt.Option)`
  - Register by constructor function.
  - Supported signatures:
    - `func(...deps) T`
    - `func(...deps) (T, errorLike)`
  - **Interface dependency arguments are not supported currently.**
- `ProvideValue[T any](value T, opts ...dirt.Option)`
  - Register a value prototype directly.
- `ProvideAs[T any, I any](opts ...dirt.Option)` **(Not implemented yet)**
  - Register type `T` as interface/base type `I`.
  - `Invoke[I]` will return `T` instance, but `InvokeAs[I]` and `InvokeAsMany[I]` stay the same.

### Resolve API

- `Invoke[T](opts ...dirt.Option) (T, error)`
  - Resolve exact type `T` from scope (with caching/reuse by scope behavior).
- `InvokeIndividual[T](opts ...dirt.Option) (T, error)`
  - Always instantiate a fresh `T` (non-shared instance).
- `InvokeAs[T](opts ...dirt.Option) (T, error)`
  - Resolve as interface/base type and return one best match.
- `InvokeAsMany[T](opts ...dirt.Option) iter.Seq2[T, error]`
  - Iterate over all matches as `T`.

### Options & Scope

- `Named(name string) dirt.Option`
  - Select/provide by name (default is empty name `""`).
- `Scoped(scope core.IScope) dirt.Option`
  - Read/write registrations and instances in the specified scope.
- `GlobalScope() *dirt.Scope`
  - Access the default global scope.
- `NewScopeWithGlobalRegistry() *dirt.Scope`
  - Create a new scope with the global registry, so it can see all global registrations but manage its own instances.

### Struct Indicators & Hook

- `dirt.InjectingGroup`
  - Mark nested struct groups for recursive field injection.
- `dirt.IPostInjectHook`
  - Implement `PostInject() error` to run post-injection initialization.

## Appendix

### Lifecycle management

```bash
go get github.com/h2570su/lifecycle
```
