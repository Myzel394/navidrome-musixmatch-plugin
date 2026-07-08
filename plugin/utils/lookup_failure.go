package utils

import (
	"errors"
	"fmt"
)

type LookupFailure struct {
	Reason     string
	Source     string
	StatusCode int
	Err        error
}

type LookupSuccess struct {
	Category string
}

type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func NewLookupFailure(reason, source string, err error) *LookupFailure {
	failure := &LookupFailure{
		Reason: reason,
		Source: source,
		Err:    err,
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		failure.StatusCode = httpErr.StatusCode
	}

	return failure
}

func NewLookupSuccess(category string) *LookupSuccess {
	return &LookupSuccess{Category: category}
}

func LookupFailureFromError(err error) *LookupFailure {
	var failure *LookupFailure
	if errors.As(err, &failure) {
		return failure
	}
	return nil
}

func (f *LookupFailure) Error() string {
	if f == nil {
		return ""
	}
	if f.Err != nil {
		return f.Err.Error()
	}
	if f.Reason != "" {
		return f.Reason
	}
	return f.ReasonValue()
}

func (f *LookupFailure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Err
}

func (f *LookupFailure) WithStatusCode(statusCode int) *LookupFailure {
	if f != nil {
		f.StatusCode = statusCode
	}
	return f
}

func (f *LookupFailure) ReasonValue() string {
	if f == nil || f.Reason == "" {
		return "unknown"
	}
	return f.Reason
}

func (f *LookupFailure) SourceValue() string {
	if f == nil || f.Source == "" {
		return "unknown"
	}
	return f.Source
}

func (s *LookupSuccess) CategoryValue() string {
	if s == nil || s.Category == "" {
		return "unknown"
	}
	return s.Category
}
