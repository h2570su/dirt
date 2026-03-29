package dirt

import "git.ttech.cc/astaroth/dirt/internal/hook"

// IPostInjectHook is an interface that can be implemented by types that want to do post-injection initialization.
type IPostInjectHook = hook.IPostInjectHook
