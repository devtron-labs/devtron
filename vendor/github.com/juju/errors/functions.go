// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package errors

import (
	stderrors "errors"
	"fmt"
	"runtime"
	"strings"
)

// New is a drop in replacement for the standard library errors module that records
// the location that the error is created.
//
// For example:
//    return errors.New("validation failed")
//
func New(message string) error {
	err := &Err{message: message}
	err.SetLocation(1)
	return err
}

// Errorf creates a new annotated error and records the location that the
// error is created.  This should be a drop in replacement for fmt.Errorf.
//
// For example:
//    return errors.Errorf("validation failed: %s", message)
//
func Errorf(format string, args ...interface{}) error {
	err := &Err{message: fmt.Sprintf(format, args...)}
	err.SetLocation(1)
	return err
}

// getLocation records the package path-qualified function name of the error at
// callDepth stack frames above the call.
func getLocation(callDepth int) (string, int) {
	rpc := make([]uintptr, 1)
	n := runtime.Callers(callDepth+2, rpc[:])
	if n < 1 {
		return "", 0
	}
	frame, _ := runtime.CallersFrames(rpc).Next()
	return frame.Function, frame.Line
}

// Trace adds the location of the Trace call to the stack.  The Cause of the
// resulting error is the same as the error parameter.  If the other error is
// nil, the result will be nil.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Trace(err)
//   }
//
func Trace(other error) error {
	//return SetLocation(other, 2)
	if other == nil {
		return nil
	}
	err := &Err{previous: other, cause: Cause(other)}
	err.SetLocation(1)
	return err
}

// Annotate is used to add extra context to an existing error. The location of
// the Annotate call is recorded with the annotations. The file, line and
// function are also recorded.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Annotate(err, "failed to frombulate")
//   }
//
func Annotate(other error, message string) error {
	if other == nil {
		return nil
	}
	err := &Err{
		previous: other,
		cause:    Cause(other),
		message:  message,
	}
	err.SetLocation(1)
	return err
}

// Annotatef is used to add extra context to an existing error. The location of
// the Annotate call is recorded with the annotations. The file, line and
// function are also recorded.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Annotatef(err, "failed to frombulate the %s", arg)
//   }
//
func Annotatef(other error, format string, args ...interface{}) error {
	if other == nil {
		return nil
	}
	err := &Err{
		previous: other,
		cause:    Cause(other),
		message:  fmt.Sprintf(format, args...),
	}
	err.SetLocation(1)
	return err
}

// DeferredAnnotatef annotates the given error (when it is not nil) with the given
// format string and arguments (like fmt.Sprintf). If *err is nil, DeferredAnnotatef
// does nothing. This method is used in a defer statement in order to annotate any
// resulting error with the same message.
//
// For example:
//
//    defer DeferredAnnotatef(&err, "failed to frombulate the %s", arg)
//
func DeferredAnnotatef(err *error, format string, args ...interface{}) {
	if *err == nil {
		return
	}
	newErr := &Err{
		message:  fmt.Sprintf(format, args...),
		cause:    Cause(*err),
		previous: *err,
	}
	newErr.SetLocation(1)
	*err = newErr
}

// Wrap changes the Cause of the error. The location of the Wrap call is also
// stored in the error stack.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       newErr := &packageError{"more context", private_value}
//       return errors.Wrap(err, newErr)
//   }
//
func Wrap(other, newDescriptive error) error {
	err := &Err{
		previous: other,
		cause:    newDescriptive,
	}
	err.SetLocation(1)
	return err
}

// Wrapf changes the Cause of the error, and adds an annotation. The location
// of the Wrap call is also stored in the error stack.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errors.Wrapf(err, simpleErrorType, "invalid value %q", value)
//   }
//
func Wrapf(other, newDescriptive error, format string, args ...interface{}) error {
	err := &Err{
		message:  fmt.Sprintf(format, args...),
		previous: other,
		cause:    newDescriptive,
	}
	err.SetLocation(1)
	return err
}

// Maskf masks the given error with the given format string and arguments (like
// fmt.Sprintf), returning a new error that maintains the error stack, but
// hides the underlying error type.  The error string still contains the full
// annotations. If you want to hide the annotations, call Wrap.
func Maskf(other error, format string, args ...interface{}) error {
	if other == nil {
		return nil
	}
	err := &Err{
		message:  fmt.Sprintf(format, args...),
		previous: other,
	}
	err.SetLocation(1)
	return err
}

// Mask hides the underlying error type, and records the location of the masking.
func Mask(other error) error {
	if other == nil {
		return nil
	}
	err := &Err{
		previous: other,
	}
	err.SetLocation(1)
	return err
}

// Cause returns the cause of the given error.  This will be either the
// original error, or the result of a Wrap or Mask call.
//
// Cause is the usual way to diagnose errors that may have been wrapped by
// the other errors functions.
func Cause(err error) error {
	var diag error
	if err, ok := err.(causer); ok {
		diag = err.Cause()
	}
	if diag != nil {
		return diag
	}
	return err
}

type causer interface {
	Cause() error
}

type wrapper interface {
	// Message returns the top level error message,
	// not including the message from the Previous
	// error.
	Message() string

	// Underlying returns the Previous error, or nil
	// if there is none.
	Underlying() error
}

var (
	_ wrapper    = (*Err)(nil)
	_ Locationer = (*Err)(nil)
	_ causer     = (*Err)(nil)
)

// Details returns information about the stack of errors wrapped by err, in
// the format:
//
// 	[{filename:99: error one} {otherfile:55: cause of error one}]
//
// This is a terse alternative to ErrorStack as it returns a single line.
func Details(err error) string {
	if err == nil {
		return "[]"
	}
	var s []byte
	s = append(s, '[')
	for {
		s = append(s, '{')
		if err, ok := err.(Locationer); ok {
			file, line := err.Location()
			if file != "" {
				s = append(s, fmt.Sprintf("%s:%d", file, line)...)
				s = append(s, ": "...)
			}
		}
		if cerr, ok := err.(wrapper); ok {
			s = append(s, cerr.Message()...)
			err = cerr.Underlying()
		} else {
			s = append(s, err.Error()...)
			err = nil
		}
		s = append(s, '}')
		if err == nil {
			break
		}
		s = append(s, ' ')
	}
	s = append(s, ']')
	return string(s)
}

// ErrorStack returns a string representation of the annotated error. If the
// error passed as the parameter is not an annotated error, the result is
// simply the result of the Error() method on that error.
//
// If the error is an annotated error, a multi-line string is returned where
// each line represents one entry in the annotation stack. The full filename
// from the call stack is used in the output.
//
//     first error
//     github.com/juju/errors/annotation_test.go:193:
//     github.com/juju/errors/annotation_test.go:194: annotation
//     github.com/juju/errors/annotation_test.go:195:
//     github.com/juju/errors/annotation_test.go:196: more context
//     github.com/juju/errors/annotation_test.go:197:
func ErrorStack(err error) string {
	return strings.Join(errorStack(err), "\n")
}

func errorStack(err error) []string {
	if err == nil {
		return nil
	}

	// We want the first error first
	var lines []string
	for {
		var buff []byte
		if err, ok := err.(Locationer); ok {
			file, line := err.Location()
			// Strip off the leading GOPATH/src path elements.
			if file != "" {
				buff = append(buff, fmt.Sprintf("%s:%d", file, line)...)
				buff = append(buff, ": "...)
			}
		}
		if cerr, ok := err.(wrapper); ok {
			message := cerr.Message()
			buff = append(buff, message...)
			// If there is a cause for this error, and it is different to the cause
			// of the underlying error, then output the error string in the stack trace.
			var cause error
			if err1, ok := err.(causer); ok {
				cause = err1.Cause()
			}
			err = cerr.Underlying()
			if cause != nil && !sameError(Cause(err), cause) {
				if message != "" {
					buff = append(buff, ": "...)
				}
				buff = append(buff, cause.Error()...)
			}
		} else {
			buff = append(buff, err.Error()...)
			err = nil
		}
		lines = append(lines, string(buff))
		if err == nil {
			break
		}
	}
	// reverse the lines to get the original error, which was at the end of
	// the list, back to the start.
	var result []string
	for i := len(lines); i > 0; i-- {
		result = append(result, lines[i-1])
	}
	return result
}

// Unwrap is a proxy for the Unwrap function in Go's standard `errors` library
// (pkg.go.dev/errors).
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// Is is a proxy for the Is function in Go's standard `errors` library
// (pkg.go.dev/errors).
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// HasType is a function wrapper around AsType dropping the where return value
// from AsType() making a function that can be used like this:
//
//  return HasType[*MyError](err)
//
// Or
//
//  if HasType[*MyError](err) {}
func HasType[T error](err error) bool {
	_, rval := AsType[T](err)
	return rval
}

// As is a proxy for the As function in Go's standard `errors` library
// (pkg.go.dev/errors).
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

// AsType is a convenience method for checking and getting an error from within
// a chain that is of type T. If no error is found of type T in the chain the
// zero value of T is returned with false. If an error in the chain implementes
// As(any) bool then it's As method will be called if it's type is not of type T.

// AsType finds the first error in err's chain that is assignable to type T, and
// if a match is found, returns that error value and true. Otherwise, it returns
// T's zero value and false.
//
// AsType is equivalent to errors.As, but uses a type parameter and returns
// the target, to avoid having to define a variable before the call. For
// example, callers can replace this:
//
//  var pathError *fs.PathError
//  if errors.As(err, &pathError) {
//      fmt.Println("Failed at path:", pathError.Path)
//  }
//
// With:
//
//  if pathError, ok := errors.AsType[*fs.PathError](err); ok {
//      fmt.Println("Failed at path:", pathError.Path)
//  }
func AsType[T error](err error) (T, bool) {
	for err != nil {
		if e, is := err.(T); is {
			return e, true
		}
		var res T
		if x, ok := err.(interface{ As(any) bool }); ok && x.As(&res) {
			return res, true
		}
		err = stderrors.Unwrap(err)
	}
	var zero T
	return zero, false
}

// SetLocation takes a given error and records where in the stack SetLocation
// was called from and returns the wrapped error with the location information
// set. The returned error implements the Locationer interface. If err is nil
// then a nil error is returned.
func SetLocation(err error, callDepth int) error {
	if err == nil {
		return nil
	}

	return newLocationError(err, callDepth)
}

// fmtNoop provides an internal type for wrapping errors so they won't be
// printed in fmt type commands. As this type is used by the Hide function it's
// expected that error not be nil.
type fmtNoop struct {
	error
}

// Format implements the fmt.Formatter interface so that the error wrapped by
// fmtNoop will not be printed.
func (*fmtNoop) Format(_ fmt.State, r rune) {}

// Is implements errors.Is. It useful for us to be able to check if an error
// chain has fmtNoop for formatting purposes.
func (f *fmtNoop) Is(err error) bool {
	_, is := err.(*fmtNoop)
	return is
}

// Unwrap implements the errors.Unwrap method returning the error wrapped by
// fmtNoop.
func (f *fmtNoop) Unwrap() error {
	return f.error
}

// Hide takes an error and silences it's error string from appearing in fmt
// like
func Hide(err error) error {
	if err == nil {
		return nil
	}
	return &fmtNoop{err}
}
