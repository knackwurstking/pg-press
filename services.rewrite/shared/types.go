package shared

import (
	"database/sql"
	"fmt"
)

type Setup struct {
	// Contains configuration parameters for the service(s)
	EnableSQL      bool   `json:"enable_sql"`
	DriverName     string `json:"driver_name"`
	DataSourceName string `json:"data_source_name"`

	// DB is the database connection instance if SQL is enabled
	DB *sql.DB `json:"-"`
}

type EntityID int64

func (id EntityID) String() string {
	return fmt.Sprintf("%d", id)
}

type UnixMilly int64

func (u UnixMilly) String() string {
	return fmt.Sprintf("%d", u)
}
