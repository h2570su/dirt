package dirt

import (
	"github.com/h2570su/dirt/internal/provide/bystruct"
)

// Injectable is an indicator to be embedded in struct to be provided by ProvideStruct.
type Injectable = bystruct.Injectable

// Subclass is an indicator to be embedded in struct to indicate it's a subclass/folder of the parent struct, which lets the dependencies inside it to be automatically resolved and injected by ProvideStruct.
type Subclass = bystruct.Subclass
