package dirt

import (
	"git.ttech.cc/astaroth/dirt/internal"
)

type iInjectable interface {
	dirtInjectable()
}

type Injectable struct{ _ internal.Sentinel }

func (Injectable) dirtInjectable() {}

type iSubclass interface {
	dirtSubclass()
}

type Subclass struct{ _ internal.Sentinel }

func (Subclass) dirtSubclass() {}
