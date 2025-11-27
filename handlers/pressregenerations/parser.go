package pressregenerations

import (
	"time"

	"github.com/labstack/echo/v4"
)

type RegenerationsFormData struct {
	Started   time.Time
	Completed time.Time
}

func ParseFormRegenerationsPage(c echo.Context) RegenerationsFormData {
	// TODO: ...

	return RegenerationsFormData{
		//Started: ,
		//Completed: ,
	}
}
