package urlb

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/shared"
)

func MetalSheetDelete(metalSheetID shared.EntityID) templ.SafeURL {
	return BuildURLWithParams("/metal-sheets/delete", map[string]string{
		"id": fmt.Sprintf("%d", metalSheetID),
	})
}
