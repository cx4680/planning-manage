package opserror

import (
	xerrors "errors"
	"sync"

	"github.com/pkg/errors"
)

var (
	messages sync.Map                // map[int]string
	codes    = map[string]struct{}{} // register codes.
	lock     = sync.Mutex{}
)

var (
	CodeNotFoundErr = xerrors.New("code not found")
)

// Register register ecode message map.
func Register(cm map[string]string) {
	for k, v := range cm {
		messages.Store(k, v)
	}
}

// New new a ecode.Codes by int value.
// NOTE: ecode must unique in global, the New will check repeat and then panic.
func New(e string, httpCode int) Code {
	return add(e)
}

// Error returns a  ecode.Codes and register associated ecode message
// NOTE: Error codes and messages should be kept together.
// ecode must unique in global, the Error will check repeat and then panic.
func Error(e string, msg string) Code {
	code := add(e)
	Register(map[string]string{
		e: msg,
	})
	return code
}
func add(e string) Code {
	lock.Lock()
	if _, ok := codes[e]; ok {
		//fmt.Printf("ecode: %d already exist \n", e)
	}
	codes[e] = struct{}{}
	lock.Unlock()
	return String(e)
}

// Codes ecode error interface which has a code & message.
type Codes interface {
	// Error return Code in string form
	Error() string
	// Code get error code.
	Code() string
	// HTTPCode http code
	HTTPCode() int
	// Message get code message.
	Message() string
	// Equal for compatible.
	Equal(error) bool
}

// A Code is an int error code spec.
type Code string

func (e Code) Error() string {
	return string(e)
}

// Code return error code
func (e Code) Code() string { return string(e) }

// Message return error message
func (e Code) Message() string {
	v, ok := messages.Load(e.Code())
	if !ok {
		return e.Error()
	}
	ret, _ := v.(string)
	return ret
}

func (e Code) HTTPCode() int {
	return 0
}

// Equal for compatible.
func (e Code) Equal(err error) bool { return EqualError(e, err) }

// A Code is an int error code spec.
type SpecialError struct {
	message  string
	code     string
	httpCode int
}

func AddSpecialError(code string, msg string, httpCode int) SpecialError {
	return SpecialError{
		message:  msg,
		code:     code,
		httpCode: httpCode,
	}
}

func (e SpecialError) Error() string {
	return e.message
}

// Code return error code
func (e SpecialError) Code() string { return string(e.code) }

// Message return error message
func (e SpecialError) Message() string {
	return e.message
}

func (e SpecialError) HTTPCode() int {
	return e.httpCode
}

// Equal for compatible.
func (e SpecialError) Equal(err error) bool { return EqualError(e, err) }

// TopCode returns the first Code object if any code type error
// in err's chain
// Otherwise, returns CodeNotFoundErr
func TopCode(err error) (Codes, error) {
	var (
		u  Codes
		ok bool
	)
	for {
		if err == nil {
			break
		}
		if u, ok = err.(Codes); ok {
			return u, nil
		}
		err = xerrors.Unwrap(err)
	}
	return nil, CodeNotFoundErr
}

// String parse code string to error.
func String(e string) Code {
	return Code(e)
}

// Cause cause from error to ecode.
func Cause(e error) Codes {
	if e == nil {
		return Success
	}
	ec, ok := errors.Cause(e).(Codes)
	if ok {
		return ec
	}

	return String(e.Error())
}

// Equal equal a and b by code int.
func Equal(a, b Codes) bool {
	if a == nil {
		a = String("")
	}
	if b == nil {
		b = String("")
	}
	return a.Code() == b.Code()
}

// EqualError equal error
func EqualError(code Codes, err error) bool {
	return Cause(err).Code() == code.Code()
}
