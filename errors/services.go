package errors

import "fmt"

type DBType string

const (
	DBTypeValidation DBType = "validation"
	DBTypeScan       DBType = "scan"
	DBTypeNotFound   DBType = "not found"
	DBTypeSelect     DBType = "SELECT"
	DBTypeInsert     DBType = "INSERT"
	DBTypeUpdate     DBType = "UPDATE"
	DBTypeDelete     DBType = "DELETE"
)

type DBError struct {
	Err error
	Typ DBType
}

func (e *DBError) Error() string {
	return fmt.Sprintf("database %s error: %v", e.Typ, e.Err)
}

func NewDBError(err error, typ DBType) *DBError {
	return &DBError{
		Err: err,
		Typ: typ,
	}
}
