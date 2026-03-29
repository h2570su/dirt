package dirt

import (
	"git.ttech.cc/astaroth/dirt/internal"
)

type IInjectable interface {
	dirtInjectable()
}

type Injectable struct{ _ internal.Sentinel }

func (Injectable) dirtInjectable() {}

type ISubclass interface {
	IInjectable
	dirtSubclass()
}

type Subclass struct{ _ internal.Sentinel }

func (Subclass) dirtSubclass() {}
