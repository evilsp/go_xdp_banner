package common

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatusError interface {
	Status() *status.Status
}

func HandleError(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(StatusError); ok {
		return e.Status().Err()
	}

	return status.New(codes.Unavailable, err.Error()).Err()
}

func InvalidArgumentError(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}
