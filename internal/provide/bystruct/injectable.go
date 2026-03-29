package bystruct

import (
	"git.ttech.cc/astaroth/dirt/internal"
)

type IInjectable interface {
	dirtInjectable()
}

// Injectable is an indicator to be embedded in struct to be provided by ProvideStruct.
type Injectable struct{ _ internal.Sentinel }

func (Injectable) dirtInjectable() {}

type ISubclass interface {
	dirtSubclass()
}

// Subclass is an indicator to be embedded in struct to indicate it's a subclass/folder of the parent struct, which lets the dependencies inside it to be automatically resolved and injected by ProvideStruct.
type Subclass struct{ _ internal.Sentinel }

func (Subclass) dirtSubclass() {}
