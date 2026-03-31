package dirt

import "github.com/h2570su/dirt/internal/hook"

// IPostInjectHook is an interface that can be implemented by types that want to do post-injection initialization.
type IPostInjectHook = hook.IPostInjectHook
