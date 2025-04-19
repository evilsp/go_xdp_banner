package errors

import (
	"errors"
	"testing"
)

func TestAppError(t *testing.T) {
	err := NewInputError("name is required")

	err2 := &AppError{}
	t.Log(errors.As(err, &err2))
	t.Log(err2)

}
