package errors

import (
	"errors"
	"fmt"
)

func Wrap(err error, format string, a ...any) error {
	msg := ""
	if format != "" {
		msg = fmt.Sprintf(format, a...)
	}

	if msg == "" {
		return err
	}

	if err == nil {
		return errors.New(msg)
	}

	// Format the wrapped error with a concise message that starts with lowercase
	return fmt.Errorf("%s: %v", msg, err)
}
