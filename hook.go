package dirt

import "github.com/h2570su/dirt/internal/hook"

type (
	// IPostInjectHookE is an interface that can be implemented by types that want to do post-injection initialization.
	IPostInjectHookE = hook.IPostInjectHookE
	// IPostInjectHook is an interface that can be implemented by types that want to do post-injection initialization.
	IPostInjectHook = hook.IPostInjectHook
)
