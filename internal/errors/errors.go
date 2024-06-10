package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NetworkError struct {
	err  error
	code codes.Code
}

func (n *NetworkError) Error() string {
	if n.err != nil {
		return n.err.Error()
	}
	return ""
}

func NewNetworkError(code codes.Code, message string) *NetworkError {
	n := &NetworkError{
		err:  errors.New(message),
		code: code,
	}
	return n
}

func (n *NetworkError) ToGRPCError() error {
	return status.Error(n.code, n.Error())
}
