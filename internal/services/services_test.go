package services

import (
	"errors"
	"testing"
)

func TestNotFoundErrorDefaultsResourceName(t *testing.T) {
	err := NotFoundError{}
	if err.Error() != "resource not found" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestNotFoundErrorTrimsResourceName(t *testing.T) {
	err := NotFoundError{Resource: "  category  "}
	if err.Error() != "category not found" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestValidationErrorTrimsMessage(t *testing.T) {
	err := ValidationError{Message: "  invalid input  "}
	if err.Error() != "invalid input" {
		t.Fatalf("unexpected validation message: %s", err.Error())
	}
}

func TestIsNotFoundMatchesWrappedError(t *testing.T) {
	wrapped := errors.New("prefix: " + NotFoundError{Resource: "product"}.Error())
	if IsNotFound(wrapped) {
		t.Fatalf("plain text error should not match IsNotFound")
	}

	err := errors.Join(errors.New("wrapper"), NotFoundError{Resource: "product"})
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound to match wrapped not found error")
	}
}

func TestIsValidationMatchesWrappedError(t *testing.T) {
	err := errors.Join(errors.New("wrapper"), ValidationError{Message: "bad payload"})
	if !IsValidation(err) {
		t.Fatalf("expected IsValidation to match wrapped validation error")
	}
}
