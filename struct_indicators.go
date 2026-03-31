package dirt

import (
	"github.com/h2570su/dirt/internal/provide/bystruct"
)

// InjectingGroup is an indicator to be embedded in struct to indicate it's a subclass/folder of the parent struct, which lets the dependencies inside it to be automatically resolved and injected by ProvideStruct.
type InjectingGroup = bystruct.InjectingGroup
