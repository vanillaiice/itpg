package itpg

import (
	"encoding/json"
	"fmt"
	"io"
)

// Response represents a response returned by the server.
type Response struct {
	Code    int         `json:"code"`    // Internal response status code
	Message interface{} `json:"message"` // Message associated with the response
}

// String returns a string representation of the Response.
func (r *Response) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// Error returns an error representation of the Response.
// It contains the error code and the message.
func (r *Response) Error() error {
	return fmt.Errorf(r.String())
}

// WriteJSON writes a response to the specified writer
func (r *Response) WriteJSON(w io.Writer) {
	w.Write([]byte(r.String()))
}

// NewResponse creates a new Response with the given code and message.
func NewResponse(code int, message interface{}) *Response {
	return &Response{Code: code, Message: message}
}

// NewErrEmptyValueFor returns a new Response struct with an error code
// indicating an empty value, and the name of the empty value
func NewErrEmptyValueFor(s string) *Response {
	return &Response{errEmptyValue.Code, fmt.Sprintf("got empty value for %s", s)}
}

// Response code indicating successful operation.
var successCode = 2000

// Success response indicating successful operation.
var success = NewResponse(successCode, "success")

// Client-side errors
var (
	// errRegistered indicates that the user is already registered.
	errRegistered = NewResponse(4000, "already registered")
	// errNotRegistered indicates that the user is not registered.
	errNotRegistered = NewResponse(4001, "not registered")
	// errUsernameTaken indicates that the username is already taken.
	errUsernameTaken = NewResponse(4002, "username taken")
	// errLoggedIn indicates that the user is already logged in.
	errLoggedIn = NewResponse(4003, "already logged in")
	// errNotLoggedIn indicates that the user is not logged in.
	errNotLoggedIn = NewResponse(4004, "not logged in")
	// errConfirmed indicates that the user account is already confirmed.
	errConfirmed = NewResponse(4005, "already confirmed")
	// errNotConfirmed indicates that the user account is not confirmed.
	errNotConfirmed = NewResponse(4006, "not confirmed")
	// errNotConfirmedUser indicates that the user is not confirmed.
	errNotConfirmedUser = NewResponse(4007, "user not confirmed")
	// errWrongUsernamePassword indicates that the username or password is incorrect.
	errWrongUsernamePassword = NewResponse(4008, "wrong username or password")
	// errWrongConfirmationCode indicates that the confirmation code is incorrect.
	errWrongConfirmationCode = NewResponse(4009, "wrong confirmation code")
	// errInvalidCookie indicates that the cookie is invalid.
	errInvalidCookie = NewResponse(4010, "invalid cookie")
	// errExpiredCookie indicates that the cookie has expired.
	errExpiredCookie = NewResponse(4011, "expired cookie")
	// errBadRequest indicates a bad request.
	errBadRequest = NewResponse(4012, "bad request")
	// errEmptyValue indicates an empty value.
	errEmptyValue = NewResponse(4013, "empty value")
	// errCourseGraded indicates that the course is already graded.
	errCourseGraded = NewResponse(4014, "course already graded")
	// errPermissionDenied indicates that permission is denied.
	errPermissionDenied = NewResponse(4015, "permission denied")
	// errInvalidEmail indicates that the email format is invalid.
	errInvalidEmail = NewResponse(4016, "invalid email format")
	// errEmailDomainNotAllowed indicates that the email domain is not allowed.
	errEmailDomainNotAllowed = NewResponse(4017, "email domain not allowed")
	// errRequestLimitReached indicates that the user has reached the request limit
	errRequestLimitReached = NewResponse(4018, "request limit reached")
)

// Server-side errors
var (
	// errGenCode indicates an error generating the confirmation code.
	errGenCode = NewResponse(5000, "error generating confirmation code")
	// errSendMail indicates an error mailing the confirmation code.
	errSendMail = NewResponse(5001, "error mailing confirmation code")
	// errInternal indicates an internal error.
	errInternal = NewResponse(5002, "internal error")
)
