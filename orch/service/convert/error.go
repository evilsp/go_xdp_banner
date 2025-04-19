package convert

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrInvalidField struct {
	field string
	msg   string
}

func (e ErrInvalidField) Error() string {
	return fmt.Sprintf("invalid field %s: %s", e.field, e.msg)
}

func (e ErrInvalidField) Field() string {
	return e.field
}

func (e ErrInvalidField) Msg() string {
	return e.msg
}

func (e ErrInvalidField) Is(target error) bool {
	_, ok := target.(ErrInvalidField)
	return ok
}

func (e ErrInvalidField) Status() *status.Status {
	st := status.New(codes.InvalidArgument, "invalid field")
	detail := &errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       e.field,
				Description: e.msg,
			},
		},
	}
	stWithDetails, _ := st.WithDetails(detail)
	return stWithDetails
}

func NewErrInvalidField(field, msg string) *ErrInvalidField {
	return &ErrInvalidField{field: field, msg: msg}
}
