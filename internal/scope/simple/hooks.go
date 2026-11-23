package simple

import (
	"reflect"
	"sync"

	"github.com/h2570su/dirt/core"
)

var _ core.IHooks = &Hooks{}

type Hooks struct {
	lock sync.RWMutex

	onInjecting core.HookFnOnInjecting
	onCtorBuilt core.HookFnCtorBuilt
}

func (s *Hooks) HookOnInjecting(hookFn core.HookFnOnInjecting) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ori := s.onInjecting
	if ori == nil {
		s.onInjecting = hookFn
		return
	}
	s.onInjecting = func(from reflect.Value, dstT reflect.Type) reflect.Value {
		from = ori(from, dstT)
		return hookFn(from, dstT)
	}
}

func (s *Hooks) OnInjecting(from reflect.Value, dstT reflect.Type) reflect.Value {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.onInjecting == nil {
		return from
	}
	return s.onInjecting(from, dstT)
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
