package simple

import (
	"reflect"
	"sync"

	"github.com/h2570su/dirt/core"
)

var _ core.IHooks = &Hooks{}

type Hooks struct {
	lock sync.RWMutex

	onInjecting []core.HookOnInjecting
	onCtorBuilt core.HookFnCtorBuilt
}

func (s *Hooks) HookOnInjecting(hookFn core.HookOnInjecting) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.onInjecting = append(s.onInjecting, hookFn)
}

func (s *Hooks) OnInjecting(from reflect.Value, dstT reflect.Type) reflect.Value {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if len(s.onInjecting) == 0 {
		return from
	}
	for _, hook := range s.onInjecting {
		if from.Type() != hook.T {
			continue
		}
		from = hook.Fn(from, dstT)
	}
	return from
}

func (s *Hooks) HookOnCtorBuilt(hookFn core.HookFnCtorBuilt) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ori := s.onCtorBuilt
	if ori == nil {
		s.onCtorBuilt = hookFn
		return
	}
	s.onCtorBuilt = func(t reflect.Type, from core.Ctor) core.Ctor {
		from = ori(t, from)
		return hookFn(t, from)
	}
}

func (s *Hooks) OnCtorBuilt(t reflect.Type, from core.Ctor) core.Ctor {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.onCtorBuilt == nil {
		return from
	}
	return s.onCtorBuilt(t, from)
}
