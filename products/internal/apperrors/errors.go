package apperrors

import (
	"errors"
	"fmt"
)

type ErrorNotFound struct {
	ID string
}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf("product with id %s not found", e.ID)
}

func IsNotFoundError(err error) bool {
	var notFoundErr *ErrorNotFound
	return errors.As(err, &notFoundErr)
}
