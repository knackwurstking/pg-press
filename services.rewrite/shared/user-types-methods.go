package shared

import "github.com/knackwurstking/pg-press/errors"

func (e *User) Validate() *errors.ValidationError

func (e *Cookie) Validate() *errors.ValidationError

func (e *Session) Validate() *errors.ValidationError
