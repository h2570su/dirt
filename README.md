# dirt is a DI framework for Go 1.22+, inspired by samber/do/v2

DIRT stands for "Dependency Injection powered by Reflection and focus on Type itself using generic".

## Why

I like the design of samber/do/v2, but in my experience with it, I found some limitations that make it less convenient to use in certain scenarios:

1. Hook doesn't give access to the instance.
1.1. No way to do post-inject initialization, which is common for struct with non-injectable fields after InjectStruct(), which I really like to use InjectStruct for field injection.
1.1.1. Even the Provide(NewXXX(i do.Injector)) is enough for custom injection logic, but it's annoying to write a lot of iferr(long) or MustInvoke(panic) in the ctor.
2. All Provide() must be hand-written(because of golang generics limitation), cannot be provided on run-time, which is useful for some dynamic initialization, e.g. based on config file or env vars.

## Thoughts

The inspiration of golang at struct is zero-value initialization, which doesn't require any ctor, even New()... pattern is common.
So I want to use injection by field as the major way, and ctor is just a supplement for some special cases, e.g. when the struct is not injectable, or when the struct is from external package and cannot be modified.

## Target

Final Target of this project is:

1. Global type registration:
1.1. `ProvideStruct[Type]()`, I hate ctor so don't want resolve deps by ctor, which is cannot be build in run-time.
1.2. `var _ = ProvideStruct[Type]()`, shorter then `func init() { do.ProvideStruct[Type]() }`
1.2.1 `ProvideStruct[Type]()` can restrict the struct to embedded a specific struct `dirt.Injectable` with a unexported method, by generic constraint.
1.2.2 `ProvideStruct[Type]()` support nested struct or *struct, by checking `dirt.Injectable`
1.3. `PreannounceInterface[Interface]()`, for faster invoke/inject when starting of invoke-chain, don't resolve on the fly.
1.3.1 Refresh implementation table on every provide/preannounce, ensure the execution order of provide/preannounce won't affect the result.
2. Post-inject initialization: if the struct implements IPostInject, PostInject() will be called after all field injections are done.
3. Inject by field tag/by Ctor, both supported. Ctor one may be `ProvideCtor(func(deps...) (Type, error))`
4. DI container majorly uses global scope, custom scope is also supported
4.1 `ProvideXXX(...Scope)`/`PreannounceInterface(...Scope)`, if Scope is not provided, use global scope by default, otherwise provide to all given scopes
4.2. `dirt.GlobalScope` is the global scope
4.3. `InvokeXXX(Scope)`, if Scope is not provided, use global scope by default, otherwise invoke from the given scope.
5. Exportable API, useful for lifecycle management, dependency graph visualization, etc.
5.1. Return all registered types, their dependencies, etc.
5.2. Return all invoked instances in `any`, with in invoke order, for lifecycle management maybe.
6. Initialization of external resources in ctor/PostInject is not recommended, initialization logic should be simplified as much as possible.
6.1. These operations are better to be done by lifecycle management. `Startup(ctx)error`/`Shutdown(ctx)error`
7. Struct tag `dirt` for field injection customization

## Features

### Tag

Fields don't hold a tag will be ignored by default.

- `dirt:"-"` to ignore the field. (even it's unnecessary)
- `dirt:"name:xxx"` to specify the name of the dependency, default is anonymous `""`.
- `dirt:"optional"` to specify the dependency is optional (default is required), if the dependency is not found or failed to invoke, let it as-is.
- `dirt:"individual"` to specify the dependency should be resolved as individual instance, which means it won't be shared with other dependencies.
- Using `,` to separate multiple options, e.g. `dirt:"name:xxx,optional"`.

## Appendix

Future work:

- Lifecycle management package
- - Support `Startup(ctx)error`/`Shutdown(ctx)error` or `Run(ctx)error`
- - As a plugin of dirt
