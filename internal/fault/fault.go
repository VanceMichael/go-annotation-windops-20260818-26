package fault

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeInvalid      Code = "invalid"
	CodeNotFound     Code = "not_found"
	CodeConflict     Code = "conflict"
	CodePrecondition Code = "precondition_failed"
	CodeCapacity     Code = "capacity_exceeded"
	CodeUnavailable  Code = "unavailable"
	CodeCanceled     Code = "canceled"
	CodeInternal     Code = "internal"
)

type Error struct {
	Code      Code
	Operation string
	Message   string
	Cause     error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Operation == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Operation, e.Message)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func New(code Code, operation, message string) error {
	return &Error{Code: code, Operation: operation, Message: message}
}

func Wrap(code Code, operation, message string, cause error) error {
	if cause == nil {
		return nil
	}
	return &Error{Code: code, Operation: operation, Message: message, Cause: cause}
}

func CodeOf(err error) Code {
	if err == nil {
		return ""
	}
	var target *Error
	if errors.As(err, &target) {
		return target.Code
	}
	return CodeInternal
}

func IsCode(err error, code Code) bool { return CodeOf(err) == code }
