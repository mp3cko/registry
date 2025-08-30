package reg

import "fmt"

// To check for errors use errors.Is(err), don't use direct comparison(==) as they are wrapped
var (
	ErrNotFound            = fmt.Errorf("not found")
	ErrNotUniqueType       = fmt.Errorf("type not unique")
	ErrNotUniqueName       = fmt.Errorf("name not unique")
	ErrNotSupported        = fmt.Errorf("not supported")
	ErrBadOption           = fmt.Errorf("bad option")
	ErrAccessibilityTooLow = fmt.Errorf("accessibility too low")
	ErrNamednessTooLow     = fmt.Errorf("namedness too low")
)
